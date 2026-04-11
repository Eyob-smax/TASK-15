package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
)

// CoachStore implements store.CoachRepository using a pgx connection pool.
type CoachStore struct {
	pool *pgxpool.Pool
}

// NewCoachStore creates a new CoachStore backed by the given pool.
func NewCoachStore(pool *pgxpool.Pool) *CoachStore {
	return &CoachStore{pool: pool}
}

// Create inserts a new coach record.
func (s *CoachStore) Create(ctx context.Context, coach *domain.Coach) error {
	const q = `
		INSERT INTO coaches (id, user_id, location_id, specialization, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := s.pool.Exec(ctx, q,
		coach.ID, coach.UserID, coach.LocationID, coach.Specialization, coach.IsActive,
		coach.CreatedAt, coach.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("coach_store.Create: %w", err)
	}
	return nil
}

// GetByID retrieves a coach by UUID.
func (s *CoachStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Coach, error) {
	const q = `
		SELECT id, user_id, location_id, specialization, is_active, created_at, updated_at
		FROM coaches WHERE id = $1 LIMIT 1`
	row := s.pool.QueryRow(ctx, q, id)
	c, err := scanCoach(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("coach_store.GetByID: %w", err)
	}
	return c, nil
}

// GetByUserID retrieves the coach record associated with a user.
func (s *CoachStore) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Coach, error) {
	const q = `
		SELECT id, user_id, location_id, specialization, is_active, created_at, updated_at
		FROM coaches WHERE user_id = $1 LIMIT 1`
	row := s.pool.QueryRow(ctx, q, userID)
	c, err := scanCoach(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("coach_store.GetByUserID: %w", err)
	}
	return c, nil
}

// List returns a paginated list of coaches, optionally filtered by location.
func (s *CoachStore) List(ctx context.Context, locationID *uuid.UUID, page, pageSize int) ([]domain.Coach, int, error) {
	var total int
	var countErr error
	if locationID != nil {
		countErr = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM coaches WHERE location_id = $1`, locationID).Scan(&total)
	} else {
		countErr = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM coaches`).Scan(&total)
	}
	if countErr != nil {
		return nil, 0, fmt.Errorf("coach_store.List count: %w", countErr)
	}

	offset := (page - 1) * pageSize
	var (
		rows pgx.Rows
		err  error
	)
	if locationID != nil {
		rows, err = s.pool.Query(ctx, `
			SELECT id, user_id, location_id, specialization, is_active, created_at, updated_at
			FROM coaches WHERE location_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
			locationID, pageSize, offset)
	} else {
		rows, err = s.pool.Query(ctx, `
			SELECT id, user_id, location_id, specialization, is_active, created_at, updated_at
			FROM coaches ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			pageSize, offset)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("coach_store.List: %w", err)
	}
	defer rows.Close()

	var coaches []domain.Coach
	for rows.Next() {
		c, err := scanCoach(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("coach_store.List scan: %w", err)
		}
		coaches = append(coaches, *c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("coach_store.List rows: %w", err)
	}
	return coaches, total, nil
}

func scanCoach(row pgx.Row) (*domain.Coach, error) {
	var c domain.Coach
	err := row.Scan(
		&c.ID, &c.UserID, &c.LocationID, &c.Specialization, &c.IsActive,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Compile-time interface assertion.
var _ store.CoachRepository = (*CoachStore)(nil)
