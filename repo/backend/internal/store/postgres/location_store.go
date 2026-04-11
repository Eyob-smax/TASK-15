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

// LocationStore implements store.LocationRepository using a pgx connection pool.
type LocationStore struct {
	pool *pgxpool.Pool
}

// NewLocationStore creates a new LocationStore backed by the given pool.
func NewLocationStore(pool *pgxpool.Pool) *LocationStore {
	return &LocationStore{pool: pool}
}

// Create inserts a new location record.
func (s *LocationStore) Create(ctx context.Context, loc *domain.Location) error {
	const q = `
		INSERT INTO locations (id, name, address, timezone, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := s.pool.Exec(ctx, q,
		loc.ID, loc.Name, loc.Address, loc.Timezone, loc.IsActive, loc.CreatedAt, loc.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("location_store.Create: %w", err)
	}
	return nil
}

// GetByID retrieves a location by UUID.
func (s *LocationStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Location, error) {
	const q = `
		SELECT id, name, address, timezone, is_active, created_at, updated_at
		FROM locations WHERE id = $1 LIMIT 1`
	row := s.pool.QueryRow(ctx, q, id)
	loc, err := scanLocation(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("location_store.GetByID: %w", err)
	}
	return loc, nil
}

// List returns a paginated list of locations ordered by name.
func (s *LocationStore) List(ctx context.Context, page, pageSize int) ([]domain.Location, int, error) {
	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM locations`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("location_store.List count: %w", err)
	}

	const q = `
		SELECT id, name, address, timezone, is_active, created_at, updated_at
		FROM locations
		ORDER BY name ASC
		LIMIT $1 OFFSET $2`
	offset := (page - 1) * pageSize
	rows, err := s.pool.Query(ctx, q, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("location_store.List: %w", err)
	}
	defer rows.Close()

	var locs []domain.Location
	for rows.Next() {
		loc, err := scanLocation(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("location_store.List scan: %w", err)
		}
		locs = append(locs, *loc)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("location_store.List rows: %w", err)
	}
	return locs, total, nil
}

func scanLocation(row pgx.Row) (*domain.Location, error) {
	var loc domain.Location
	var updatedAt time.Time
	err := row.Scan(&loc.ID, &loc.Name, &loc.Address, &loc.Timezone, &loc.IsActive, &loc.CreatedAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	loc.UpdatedAt = updatedAt
	return &loc, nil
}

// Compile-time interface assertion.
var _ store.LocationRepository = (*LocationStore)(nil)
