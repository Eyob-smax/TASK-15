package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
)

// LandedCostStore implements store.LandedCostRepository using a pgx connection pool.
type LandedCostStore struct {
	pool *pgxpool.Pool
}

// NewLandedCostStore creates a new LandedCostStore backed by the given pool.
func NewLandedCostStore(pool *pgxpool.Pool) *LandedCostStore {
	return &LandedCostStore{pool: pool}
}

// Create inserts a new landed cost entry.
func (s *LandedCostStore) Create(ctx context.Context, entry *domain.LandedCostEntry) error {
	const q = `
		INSERT INTO landed_cost_entries
		    (id, item_id, purchase_order_id, po_line_id, period, cost_component,
		     raw_amount, allocated_amount, allocation_method, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := s.pool.Exec(ctx, q,
		entry.ID, entry.ItemID, entry.PurchaseOrderID, entry.POLineID, entry.Period,
		entry.CostComponent, entry.RawAmount, entry.AllocatedAmount, entry.AllocationMethod,
		entry.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("landed_cost_store.Create: %w", err)
	}
	return nil
}

// ListByItemAndPeriod returns all landed cost entries for a given item and accounting period.
func (s *LandedCostStore) ListByItemAndPeriod(ctx context.Context, itemID uuid.UUID, period string) ([]domain.LandedCostEntry, error) {
	const q = `
		SELECT id, item_id, purchase_order_id, po_line_id, period, cost_component,
		       raw_amount, allocated_amount, allocation_method, created_at
		FROM landed_cost_entries
		WHERE item_id = $1 AND period = $2
		ORDER BY created_at DESC`
	rows, err := s.pool.Query(ctx, q, itemID, period)
	if err != nil {
		return nil, fmt.Errorf("landed_cost_store.ListByItemAndPeriod: %w", err)
	}
	defer rows.Close()
	return scanLandedCostRows(rows)
}

// ListByPOID returns all landed cost entries for a given purchase order.
func (s *LandedCostStore) ListByPOID(ctx context.Context, poID uuid.UUID) ([]domain.LandedCostEntry, error) {
	const q = `
		SELECT id, item_id, purchase_order_id, po_line_id, period, cost_component,
		       raw_amount, allocated_amount, allocation_method, created_at
		FROM landed_cost_entries
		WHERE purchase_order_id = $1
		ORDER BY created_at DESC`
	rows, err := s.pool.Query(ctx, q, poID)
	if err != nil {
		return nil, fmt.Errorf("landed_cost_store.ListByPOID: %w", err)
	}
	defer rows.Close()
	return scanLandedCostRows(rows)
}

func scanLandedCostRows(rows pgx.Rows) ([]domain.LandedCostEntry, error) {
	var entries []domain.LandedCostEntry
	for rows.Next() {
		var e domain.LandedCostEntry
		if err := rows.Scan(
			&e.ID, &e.ItemID, &e.PurchaseOrderID, &e.POLineID, &e.Period,
			&e.CostComponent, &e.RawAmount, &e.AllocatedAmount, &e.AllocationMethod,
			&e.CreatedAt,
		); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

// Compile-time interface assertion.
var _ store.LandedCostRepository = (*LandedCostStore)(nil)
