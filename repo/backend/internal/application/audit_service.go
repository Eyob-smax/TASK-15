package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/platform"
	"fitcommerce/internal/security"
	"fitcommerce/internal/store"
)

// AuditServiceImpl is the concrete implementation of AuditService. The audit log
// is append-only: events are written with SHA-256 hash chaining for tamper
// detection. No update or delete operations are permitted.
type AuditServiceImpl struct {
	repo store.AuditRepository
}

// NewAuditService creates an AuditServiceImpl backed by the given repository.
func NewAuditService(repo store.AuditRepository) *AuditServiceImpl {
	return &AuditServiceImpl{repo: repo}
}

// Log appends a new event to the audit trail. It:
//  1. Fetches the integrity hash of the last event to maintain the hash chain.
//  2. Strips sensitive fields from details via SafeDetails.
//  3. Constructs and hashes a new AuditEvent.
//  4. Persists it via the repository.
func (s *AuditServiceImpl) Log(
	ctx context.Context,
	eventType, entityType string,
	entityID, actorID uuid.UUID,
	details map[string]interface{},
) error {
	previousHash, err := s.repo.GetLatestHash(ctx)
	if err != nil {
		return fmt.Errorf("audit_service.Log get latest hash: %w", err)
	}

	event := security.BuildAuditEvent(
		eventType,
		entityType,
		entityID,
		actorID,
		security.SafeDetails(details),
		previousHash,
	)

	if err := s.repo.Create(ctx, &event); err != nil {
		return fmt.Errorf("audit_service.Log create: %w", err)
	}
	platform.LogAuditEvent(slog.Default(), eventType, entityType, entityID)
	return nil
}

// List returns a paginated list of audit events, optionally filtered by entity
// type and entity ID.
func (s *AuditServiceImpl) List(
	ctx context.Context,
	entityType string,
	entityID *uuid.UUID,
	page, pageSize int,
) ([]domain.AuditEvent, int, error) {
	return s.repo.List(ctx, entityType, entityID, page, pageSize)
}

// ListByEventTypes returns a paginated list of audit events filtered by the
// given set of event type strings. Used for security-event inspection such as
// login failures, lockouts, and session events.
func (s *AuditServiceImpl) ListByEventTypes(
	ctx context.Context,
	eventTypes []string,
	page, pageSize int,
) ([]domain.AuditEvent, int, error) {
	return s.repo.ListByEventTypes(ctx, eventTypes, page, pageSize)
}

// Compile-time interface assertion.
var _ AuditService = (*AuditServiceImpl)(nil)
