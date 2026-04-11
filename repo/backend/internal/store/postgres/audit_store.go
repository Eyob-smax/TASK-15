package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
)

// AuditStore implements store.AuditRepository using a pgx connection pool.
// The audit log is append-only: no UPDATE or DELETE operations are permitted.
type AuditStore struct {
	pool *pgxpool.Pool
}

// NewAuditStore creates a new AuditStore backed by the given connection pool.
func NewAuditStore(pool *pgxpool.Pool) *AuditStore {
	return &AuditStore{pool: pool}
}

// Create appends a new audit event to the log. This is the only write operation
// permitted on audit_events; the table is never updated or deleted from.
func (s *AuditStore) Create(ctx context.Context, event *domain.AuditEvent) error {
	db := executorFromContext(ctx, s.pool)
	detailsJSON, err := json.Marshal(event.Details)
	if err != nil {
		return fmt.Errorf("audit_store.Create marshal details: %w", err)
	}

	const q = `
		INSERT INTO audit_events
			(id, event_type, entity_type, entity_id, actor_id, details, integrity_hash, previous_hash, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = db.Exec(ctx, q,
		event.ID,
		event.EventType,
		event.EntityType,
		event.EntityID,
		event.ActorID,
		detailsJSON,
		event.IntegrityHash,
		event.PreviousHash,
		event.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("audit_store.Create: %w", err)
	}
	return nil
}

// GetLatestHash returns the integrity_hash of the most recently inserted audit
// event, used to chain the next event. Returns an empty string if the table is
// empty (first event has no predecessor).
func (s *AuditStore) GetLatestHash(ctx context.Context) (string, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `SELECT integrity_hash FROM audit_events ORDER BY created_at DESC LIMIT 1`
	row := db.QueryRow(ctx, q)
	var hash string
	err := row.Scan(&hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("audit_store.GetLatestHash: %w", err)
	}
	return hash, nil
}

// List returns a paginated set of audit events filtered by optional entityType
// and entityID. A blank entityType matches all types; a nil entityID matches
// all entities.
func (s *AuditStore) List(
	ctx context.Context,
	entityType string,
	entityID *uuid.UUID,
	page, pageSize int,
) ([]domain.AuditEvent, int, error) {
	db := executorFromContext(ctx, s.pool)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	args := []interface{}{}
	where := "WHERE 1=1"
	argIdx := 1

	if entityType != "" {
		where += fmt.Sprintf(" AND entity_type = $%d", argIdx)
		args = append(args, entityType)
		argIdx++
	}
	if entityID != nil {
		where += fmt.Sprintf(" AND entity_id = $%d", argIdx)
		args = append(args, *entityID)
		argIdx++
	}

	countQ := fmt.Sprintf("SELECT COUNT(*) FROM audit_events %s", where)
	var total int
	if err := db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("audit_store.List count: %w", err)
	}

	listArgs := append(args, pageSize, offset)
	dataQ := fmt.Sprintf(`
		SELECT id, event_type, entity_type, entity_id, actor_id,
		       details, integrity_hash, previous_hash, created_at
		FROM audit_events
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	rows, err := db.Query(ctx, dataQ, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("audit_store.List query: %w", err)
	}
	defer rows.Close()

	var events []domain.AuditEvent
	for rows.Next() {
		var e domain.AuditEvent
		var detailsJSON []byte
		if err := rows.Scan(
			&e.ID, &e.EventType, &e.EntityType, &e.EntityID, &e.ActorID,
			&detailsJSON, &e.IntegrityHash, &e.PreviousHash, &e.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("audit_store.List scan: %w", err)
		}
		if err := json.Unmarshal(detailsJSON, &e.Details); err != nil {
			e.Details = map[string]interface{}{}
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("audit_store.List rows: %w", err)
	}

	return events, total, nil
}

// ListByEventTypes returns a paginated set of audit events whose event_type is
// one of the provided values. It is used for security-event inspection.
// Returns nil slices (not an error) when eventTypes is empty.
func (s *AuditStore) ListByEventTypes(
	ctx context.Context,
	eventTypes []string,
	page, pageSize int,
) ([]domain.AuditEvent, int, error) {
	db := executorFromContext(ctx, s.pool)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Build a parameterized IN clause
	if len(eventTypes) == 0 {
		return nil, 0, nil
	}
	args := make([]interface{}, len(eventTypes))
	placeholders := make([]string, len(eventTypes))
	for i, et := range eventTypes {
		args[i] = et
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	inClause := strings.Join(placeholders, ", ")

	countQ := fmt.Sprintf("SELECT COUNT(*) FROM audit_events WHERE event_type IN (%s)", inClause)
	var total int
	if err := db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("audit_store.ListByEventTypes count: %w", err)
	}

	listArgs := append(args, pageSize, offset)
	dataQ := fmt.Sprintf(`
		SELECT id, event_type, entity_type, entity_id, actor_id,
		       details, integrity_hash, previous_hash, created_at
		FROM audit_events
		WHERE event_type IN (%s)
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, inClause, len(eventTypes)+1, len(eventTypes)+2)

	rows, err := db.Query(ctx, dataQ, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("audit_store.ListByEventTypes query: %w", err)
	}
	defer rows.Close()

	var events []domain.AuditEvent
	for rows.Next() {
		var e domain.AuditEvent
		var detailsJSON []byte
		if err := rows.Scan(
			&e.ID, &e.EventType, &e.EntityType, &e.EntityID, &e.ActorID,
			&detailsJSON, &e.IntegrityHash, &e.PreviousHash, &e.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("audit_store.ListByEventTypes scan: %w", err)
		}
		if err := json.Unmarshal(detailsJSON, &e.Details); err != nil {
			e.Details = map[string]interface{}{}
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("audit_store.ListByEventTypes rows: %w", err)
	}
	return events, total, nil
}

// Compile-time interface assertion.
var _ store.AuditRepository = (*AuditStore)(nil)
