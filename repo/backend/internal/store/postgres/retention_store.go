package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
)

// RetentionStore implements store.RetentionRepository using a pgx connection pool.
type RetentionStore struct {
	pool *pgxpool.Pool
}

// NewRetentionStore creates a new RetentionStore backed by the given connection pool.
func NewRetentionStore(pool *pgxpool.Pool) *RetentionStore {
	return &RetentionStore{pool: pool}
}

// Create inserts a new retention policy record.
func (s *RetentionStore) Create(ctx context.Context, policy *domain.RetentionPolicy) error {
	const q = `
		INSERT INTO retention_policies (id, entity_type, retention_days, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := s.pool.Exec(ctx, q,
		policy.ID,
		policy.EntityType,
		policy.RetentionDays,
		policy.Description,
		policy.CreatedAt,
		policy.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("retention_store.Create: %w", err)
	}
	return nil
}

// GetByEntityType returns the retention policy for the given entity type.
// Returns domain.ErrNotFound if no matching record exists.
func (s *RetentionStore) GetByEntityType(ctx context.Context, entityType string) (*domain.RetentionPolicy, error) {
	const q = `
		SELECT id, entity_type, retention_days, description, created_at, updated_at
		FROM retention_policies
		WHERE entity_type = $1
		LIMIT 1`

	row := s.pool.QueryRow(ctx, q, entityType)
	policy, err := scanRetentionPolicy(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("retention_store.GetByEntityType: %w", err)
	}
	return policy, nil
}

// List returns all retention policies ordered alphabetically by entity type.
// No pagination is applied because this table is expected to remain small.
func (s *RetentionStore) List(ctx context.Context) ([]domain.RetentionPolicy, error) {
	const q = `
		SELECT id, entity_type, retention_days, description, created_at, updated_at
		FROM retention_policies
		ORDER BY entity_type ASC`

	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("retention_store.List query: %w", err)
	}
	defer rows.Close()

	var policies []domain.RetentionPolicy
	for rows.Next() {
		policy, err := scanRetentionPolicy(rows)
		if err != nil {
			return nil, fmt.Errorf("retention_store.List scan: %w", err)
		}
		policies = append(policies, *policy)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("retention_store.List rows: %w", err)
	}
	return policies, nil
}

// Update modifies the retention_days and description of an existing policy.
func (s *RetentionStore) Update(ctx context.Context, policy *domain.RetentionPolicy) error {
	const q = `
		UPDATE retention_policies
		SET retention_days = $1, description = $2, updated_at = $3
		WHERE id = $4`

	_, err := s.pool.Exec(ctx, q,
		policy.RetentionDays,
		policy.Description,
		policy.UpdatedAt,
		policy.ID,
	)
	if err != nil {
		return fmt.Errorf("retention_store.Update: %w", err)
	}
	return nil
}

// DeleteByIDs deletes records from the given table whose id column matches any
// value in ids. It uses executorFromContext so it participates in any ambient
// transaction set by postgres.WithTransaction.
func (s *RetentionStore) DeleteByIDs(ctx context.Context, table string, ids []uuid.UUID) (int64, error) {
	db := executorFromContext(ctx, s.pool)
	q := fmt.Sprintf("DELETE FROM %s WHERE id = ANY($1)", table)
	tag, err := db.Exec(ctx, q, ids)
	if err != nil {
		return 0, fmt.Errorf("retention_store.DeleteByIDs(%s): %w", table, err)
	}
	return tag.RowsAffected(), nil
}

// scanRetentionPolicy reads a single retention_policies row into a domain struct.
func scanRetentionPolicy(row pgx.Row) (*domain.RetentionPolicy, error) {
	var p domain.RetentionPolicy
	var id uuid.UUID
	var createdAt time.Time
	var updatedAt time.Time
	err := row.Scan(
		&id,
		&p.EntityType,
		&p.RetentionDays,
		&p.Description,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.ID = id
	p.CreatedAt = createdAt
	p.UpdatedAt = updatedAt
	return &p, nil
}

// Compile-time interface assertion.
var _ store.RetentionRepository = (*RetentionStore)(nil)
