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

// Compile-time interface compliance assertions.
var (
	_ store.InventoryRepository    = (*InventoryStore)(nil)
	_ store.WarehouseBinRepository = (*InventoryStore)(nil)
)

// InventoryStore implements InventoryRepository and WarehouseBinRepository.
type InventoryStore struct {
	pool *pgxpool.Pool
}

// NewInventoryStore creates a new InventoryStore backed by the given pool.
func NewInventoryStore(pool *pgxpool.Pool) *InventoryStore {
	return &InventoryStore{pool: pool}
}

// --- InventoryRepository ---

func (s *InventoryStore) CreateSnapshot(ctx context.Context, snapshot *domain.InventorySnapshot) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		INSERT INTO inventory_snapshots (id, item_id, quantity, location_id, recorded_at)
		VALUES ($1,$2,$3,$4,$5)`
	_, err := db.Exec(ctx, q,
		snapshot.ID, snapshot.ItemID, snapshot.Quantity,
		snapshot.LocationID, snapshot.RecordedAt,
	)
	return err
}

func (s *InventoryStore) ListSnapshots(ctx context.Context, itemID *uuid.UUID, locationID *uuid.UUID) ([]domain.InventorySnapshot, error) {
	db := executorFromContext(ctx, s.pool)
	q := `SELECT id, item_id, quantity, location_id, recorded_at FROM inventory_snapshots WHERE 1=1`
	args := []interface{}{}
	idx := 1

	if itemID != nil {
		q += fmt.Sprintf(` AND item_id = $%d`, idx)
		args = append(args, *itemID)
		idx++
	}
	if locationID != nil {
		q += fmt.Sprintf(` AND location_id = $%d`, idx)
		args = append(args, *locationID)
		idx++
	}
	q += ` ORDER BY recorded_at DESC`

	rows, err := db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []domain.InventorySnapshot
	for rows.Next() {
		var sn domain.InventorySnapshot
		if err := rows.Scan(&sn.ID, &sn.ItemID, &sn.Quantity, &sn.LocationID, &sn.RecordedAt); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, sn)
	}
	return snapshots, rows.Err()
}

func (s *InventoryStore) CreateAdjustment(ctx context.Context, adj *domain.InventoryAdjustment) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		INSERT INTO inventory_adjustments (id, item_id, quantity_change, reason, created_by, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := db.Exec(ctx, q,
		adj.ID, adj.ItemID, adj.QuantityChange,
		adj.Reason, adj.CreatedBy, adj.CreatedAt,
	)
	return err
}

func (s *InventoryStore) ListAdjustments(ctx context.Context, itemID *uuid.UUID, page, pageSize int) ([]domain.InventoryAdjustment, int, error) {
	db := executorFromContext(ctx, s.pool)
	countQ := `SELECT COUNT(*) FROM inventory_adjustments WHERE 1=1`
	listQ := `SELECT id, item_id, quantity_change, reason, created_by, created_at FROM inventory_adjustments WHERE 1=1`
	args := []interface{}{}
	idx := 1

	if itemID != nil {
		filter := fmt.Sprintf(` AND item_id = $%d`, idx)
		countQ += filter
		listQ += filter
		args = append(args, *itemID)
		idx++
	}

	var total int
	if err := db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQ += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1)
	listArgs := append(args, pageSize, (page-1)*pageSize)

	rows, err := db.Query(ctx, listQ, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var adjs []domain.InventoryAdjustment
	for rows.Next() {
		var a domain.InventoryAdjustment
		if err := rows.Scan(&a.ID, &a.ItemID, &a.QuantityChange, &a.Reason, &a.CreatedBy, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		adjs = append(adjs, a)
	}
	return adjs, total, rows.Err()
}

// --- WarehouseBinRepository ---

func (s *InventoryStore) Create(ctx context.Context, bin *domain.WarehouseBin) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		INSERT INTO warehouse_bins (id, location_id, name, description, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := db.Exec(ctx, q,
		bin.ID, bin.LocationID, bin.Name, bin.Description,
		bin.CreatedAt, bin.UpdatedAt,
	)
	return err
}

func (s *InventoryStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.WarehouseBin, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `SELECT id, location_id, name, description, created_at, updated_at FROM warehouse_bins WHERE id = $1`
	var bin domain.WarehouseBin
	err := db.QueryRow(ctx, q, id).Scan(
		&bin.ID, &bin.LocationID, &bin.Name, &bin.Description,
		&bin.CreatedAt, &bin.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &bin, nil
}

func (s *InventoryStore) List(ctx context.Context, locationID *uuid.UUID, page, pageSize int) ([]domain.WarehouseBin, int, error) {
	db := executorFromContext(ctx, s.pool)
	countQ := `SELECT COUNT(*) FROM warehouse_bins WHERE 1=1`
	listQ := `SELECT id, location_id, name, description, created_at, updated_at FROM warehouse_bins WHERE 1=1`
	args := []interface{}{}
	idx := 1

	if locationID != nil {
		filter := fmt.Sprintf(` AND location_id = $%d`, idx)
		countQ += filter
		listQ += filter
		args = append(args, *locationID)
		idx++
	}

	var total int
	if err := db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQ += fmt.Sprintf(` ORDER BY name LIMIT $%d OFFSET $%d`, idx, idx+1)
	listArgs := append(args, pageSize, (page-1)*pageSize)

	rows, err := db.Query(ctx, listQ, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var bins []domain.WarehouseBin
	for rows.Next() {
		var b domain.WarehouseBin
		if err := rows.Scan(&b.ID, &b.LocationID, &b.Name, &b.Description, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, 0, err
		}
		bins = append(bins, b)
	}
	return bins, total, rows.Err()
}
