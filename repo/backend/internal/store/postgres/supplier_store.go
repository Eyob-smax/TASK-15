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

// SupplierStore implements store.SupplierRepository using a pgx connection pool.
type SupplierStore struct {
	pool *pgxpool.Pool
}

// NewSupplierStore creates a new SupplierStore backed by the given connection pool.
func NewSupplierStore(pool *pgxpool.Pool) *SupplierStore {
	return &SupplierStore{pool: pool}
}

// Create inserts a new supplier record into the database.
func (s *SupplierStore) Create(ctx context.Context, supplier *domain.Supplier) error {
	const q = `
		INSERT INTO suppliers
		    (id, name, contact_name, contact_email, contact_phone, address, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`

	_, err := s.pool.Exec(ctx, q,
		supplier.ID,
		supplier.Name,
		supplier.ContactName,
		supplier.ContactEmail,
		supplier.ContactPhone,
		supplier.Address,
		supplier.IsActive,
		supplier.CreatedAt,
		supplier.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("supplier_store.Create: %w", err)
	}
	return nil
}

// GetByID retrieves a supplier by its UUID. Returns domain.ErrNotFound if absent.
func (s *SupplierStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Supplier, error) {
	const q = `
		SELECT id, name, contact_name, contact_email, contact_phone, address, is_active, created_at, updated_at
		FROM suppliers
		WHERE id = $1
		LIMIT 1`

	row := s.pool.QueryRow(ctx, q, id)
	supplier, err := scanSupplier(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("supplier_store.GetByID: %w", err)
	}
	return supplier, nil
}

// List returns a paginated list of all suppliers ordered by created_at DESC.
func (s *SupplierStore) List(ctx context.Context, page, pageSize int) ([]domain.Supplier, int, error) {
	const countQ = `SELECT COUNT(*) FROM suppliers`

	var total int
	if err := s.pool.QueryRow(ctx, countQ).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("supplier_store.List: %w", err)
	}

	const q = `
		SELECT id, name, contact_name, contact_email, contact_phone, address, is_active, created_at, updated_at
		FROM suppliers
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	offset := (page - 1) * pageSize
	rows, err := s.pool.Query(ctx, q, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("supplier_store.List: %w", err)
	}
	defer rows.Close()

	var suppliers []domain.Supplier
	for rows.Next() {
		sup, err := scanSupplier(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("supplier_store.List: %w", err)
		}
		suppliers = append(suppliers, *sup)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("supplier_store.List: %w", err)
	}
	return suppliers, total, nil
}

// Update persists all mutable fields of a supplier record.
func (s *SupplierStore) Update(ctx context.Context, supplier *domain.Supplier) error {
	const q = `
		UPDATE suppliers
		SET name          = $1,
		    contact_name  = $2,
		    contact_email = $3,
		    contact_phone = $4,
		    address       = $5,
		    is_active     = $6,
		    updated_at    = $7
		WHERE id = $8`

	_, err := s.pool.Exec(ctx, q,
		supplier.Name,
		supplier.ContactName,
		supplier.ContactEmail,
		supplier.ContactPhone,
		supplier.Address,
		supplier.IsActive,
		time.Now().UTC(),
		supplier.ID,
	)
	if err != nil {
		return fmt.Errorf("supplier_store.Update: %w", err)
	}
	return nil
}

// scanSupplier reads a single supplier row from a pgx row scanner.
func scanSupplier(row pgx.Row) (*domain.Supplier, error) {
	var sup domain.Supplier
	err := row.Scan(
		&sup.ID,
		&sup.Name,
		&sup.ContactName,
		&sup.ContactEmail,
		&sup.ContactPhone,
		&sup.Address,
		&sup.IsActive,
		&sup.CreatedAt,
		&sup.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sup, nil
}

// Compile-time interface assertion.
var _ store.SupplierRepository = (*SupplierStore)(nil)
