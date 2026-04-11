package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
)

// RetentionServiceImpl implements RetentionService.
type RetentionServiceImpl struct {
	retentionRepo store.RetentionRepository
	audit         AuditService
	pool          *pgxpool.Pool
}

// NewRetentionService creates a RetentionServiceImpl backed by the given repository and audit service.
func NewRetentionService(
	retentionRepo store.RetentionRepository,
	auditSvc AuditService,
	pool *pgxpool.Pool,
) *RetentionServiceImpl {
	return &RetentionServiceImpl{
		retentionRepo: retentionRepo,
		audit:         auditSvc,
		pool:          pool,
	}
}

// GetByEntityType retrieves the retention policy for the given entity type.
func (s *RetentionServiceImpl) GetByEntityType(ctx context.Context, entityType string) (*domain.RetentionPolicy, error) {
	return s.retentionRepo.GetByEntityType(ctx, entityType)
}

// List returns all retention policies.
func (s *RetentionServiceImpl) List(ctx context.Context) ([]domain.RetentionPolicy, error) {
	return s.retentionRepo.List(ctx)
}

// Update upserts a retention policy. If no policy exists for the entity type, a new one
// is created; otherwise the existing policy's fields are updated.
func (s *RetentionServiceImpl) Update(ctx context.Context, policy *domain.RetentionPolicy) error {
	existing, err := s.retentionRepo.GetByEntityType(ctx, policy.EntityType)
	if errors.Is(err, domain.ErrNotFound) {
		policy.ID = uuid.New()
		now := time.Now().UTC()
		policy.CreatedAt = now
		policy.UpdatedAt = now
		if err := s.retentionRepo.Create(ctx, policy); err != nil {
			return fmt.Errorf("retention_service.Update create: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("retention_service.Update get: %w", err)
	} else {
		existing.RetentionDays = policy.RetentionDays
		if policy.Description != "" {
			existing.Description = policy.Description
		}
		existing.UpdatedAt = time.Now().UTC()
		if err := s.retentionRepo.Update(ctx, existing); err != nil {
			return fmt.Errorf("retention_service.Update: %w", err)
		}
		// Ensure policy.ID is set to the existing record's ID for the audit event.
		policy.ID = existing.ID
		policy.CreatedAt = existing.CreatedAt
		policy.UpdatedAt = existing.UpdatedAt
	}

	if err := s.audit.Log(ctx, "retention.policy.updated", "retention_policy", policy.ID, domain.SystemActorID, map[string]interface{}{
		"entity_type":    policy.EntityType,
		"retention_days": policy.RetentionDays,
	}); err != nil {
		slog.Default().Warn("audit log failed", "event", "retention.policy.updated", "error", err)
	}

	return nil
}

// purgeTargets maps entity_type values to the SQL table and timestamp column
// that determine record age for deletion. Only immutable/archivable records
// are eligible for hard deletion; core business entities are excluded.
var purgeTargets = map[string]struct{ table, column string }{
	"sessions":            {table: "sessions", column: "absolute_expires_at"},
	"captcha":             {table: "captcha_challenges", column: "created_at"},
	"financial_records":   {table: "orders", column: "created_at"},
	"procurement_records": {table: "purchase_orders", column: "created_at"},
	"access_logs":         {table: "sessions", column: "absolute_expires_at"},
}

// RunCleanup enforces each active retention policy by deleting records older
// than the configured retention window. Auditable purge events are recorded.
func (s *RetentionServiceImpl) RunCleanup(ctx context.Context) error {
	policies, err := s.retentionRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("retention_service.RunCleanup list: %w", err)
	}

	now := time.Now().UTC()
	for _, policy := range policies {
		target, ok := purgeTargets[policy.EntityType]
		if !ok {
			slog.Default().Info("retention_service.RunCleanup: no purge target for entity type (skipped)",
				"entity_type", policy.EntityType)
			continue
		}

		cutoff := now.AddDate(0, 0, -policy.RetentionDays)
		selectQ := fmt.Sprintf("SELECT id FROM %s WHERE %s < $1", target.table, target.column)
		rows, err := s.pool.Query(ctx, selectQ, cutoff)
		if err != nil {
			slog.Default().Error("retention_service.RunCleanup: selection failed",
				"entity_type", policy.EntityType, "error", err)
			continue
		}
		var recordIDs []uuid.UUID
		for rows.Next() {
			var id uuid.UUID
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				slog.Default().Error("retention_service.RunCleanup: selection scan failed",
					"entity_type", policy.EntityType, "error", err)
				recordIDs = nil
				break
			}
			recordIDs = append(recordIDs, id)
		}
		if err := rows.Err(); err != nil {
			slog.Default().Error("retention_service.RunCleanup: selection rows failed",
				"entity_type", policy.EntityType, "error", err)
			continue
		}
		rows.Close()
		if len(recordIDs) == 0 {
			continue
		}

		for _, recordID := range recordIDs {
			_ = s.audit.Log(ctx, "retention.record_deleted", target.table, recordID, domain.SystemActorID, map[string]interface{}{
				"entity_type": policy.EntityType,
				"cutoff":      cutoff.Format(time.RFC3339),
			})
		}

		deleteQ := fmt.Sprintf("DELETE FROM %s WHERE id = ANY($1)", target.table)
		tag, err := s.pool.Exec(ctx, deleteQ, recordIDs)
		if err != nil {
			slog.Default().Error("retention_service.RunCleanup: purge failed",
				"entity_type", policy.EntityType, "error", err)
			continue
		}

		_ = s.audit.Log(ctx, "retention.purged", "retention_policy", policy.ID, domain.SystemActorID, map[string]interface{}{
			"entity_type":  policy.EntityType,
			"cutoff":       cutoff.Format(time.RFC3339),
			"rows_deleted": tag.RowsAffected(),
			"record_ids":   recordIDs,
		})
	}
	return nil
}

// Compile-time interface assertion.
var _ RetentionService = (*RetentionServiceImpl)(nil)
