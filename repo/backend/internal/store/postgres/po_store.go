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

// ---------------------------------------------------------------------------
// PurchaseOrderStore
// ---------------------------------------------------------------------------

// PurchaseOrderStore implements store.PurchaseOrderRepository using a pgx pool.
type PurchaseOrderStore struct {
	pool *pgxpool.Pool
}

// NewPurchaseOrderStore creates a new PurchaseOrderStore backed by the given pool.
func NewPurchaseOrderStore(pool *pgxpool.Pool) *PurchaseOrderStore {
	return &PurchaseOrderStore{pool: pool}
}

// Create inserts a new purchase order record into the database.
func (s *PurchaseOrderStore) Create(ctx context.Context, po *domain.PurchaseOrder) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		INSERT INTO purchase_orders
		    (id, supplier_id, status, total_amount, created_by, approved_by,
		     created_at, approved_at, received_at, version)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`

	_, err := db.Exec(ctx, q,
		po.ID,
		po.SupplierID,
		string(po.Status),
		po.TotalAmount,
		po.CreatedBy,
		po.ApprovedBy,
		po.CreatedAt,
		po.ApprovedAt,
		po.ReceivedAt,
		po.Version,
	)
	if err != nil {
		return fmt.Errorf("po_store.Create: %w", err)
	}
	return nil
}

// GetByID retrieves a purchase order by its UUID. Returns domain.ErrNotFound if absent.
func (s *PurchaseOrderStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.PurchaseOrder, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `
		SELECT id, supplier_id, status, total_amount, created_by, approved_by,
		       created_at, approved_at, received_at, version
		FROM purchase_orders
		WHERE id = $1
		LIMIT 1`

	row := db.QueryRow(ctx, q, id)
	po, err := scanPO(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("po_store.GetByID: %w", err)
	}
	return po, nil
}

// List returns a paginated list of all purchase orders ordered by created_at DESC.
func (s *PurchaseOrderStore) List(ctx context.Context, page, pageSize int) ([]domain.PurchaseOrder, int, error) {
	db := executorFromContext(ctx, s.pool)
	const countQ = `SELECT COUNT(*) FROM purchase_orders`

	var total int
	if err := db.QueryRow(ctx, countQ).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("po_store.List: %w", err)
	}

	const q = `
		SELECT id, supplier_id, status, total_amount, created_by, approved_by,
		       created_at, approved_at, received_at, version
		FROM purchase_orders
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	offset := (page - 1) * pageSize
	rows, err := db.Query(ctx, q, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("po_store.List: %w", err)
	}
	defer rows.Close()

	var pos []domain.PurchaseOrder
	for rows.Next() {
		po, err := scanPO(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("po_store.List: %w", err)
		}
		pos = append(pos, *po)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("po_store.List: %w", err)
	}
	return pos, total, nil
}

// Update persists mutable fields of a purchase order record.
func (s *PurchaseOrderStore) Update(ctx context.Context, po *domain.PurchaseOrder) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		UPDATE purchase_orders
		SET status       = $1,
		    total_amount = $2,
		    approved_by  = $3,
		    approved_at  = $4,
		    received_at  = $5,
		    version      = $6
		WHERE id = $7`

	_, err := db.Exec(ctx, q,
		string(po.Status),
		po.TotalAmount,
		po.ApprovedBy,
		po.ApprovedAt,
		po.ReceivedAt,
		po.Version,
		po.ID,
	)
	if err != nil {
		return fmt.Errorf("po_store.Update: %w", err)
	}
	return nil
}

// scanPO reads a single purchase_order row from a pgx row scanner.
func scanPO(row pgx.Row) (*domain.PurchaseOrder, error) {
	var po domain.PurchaseOrder
	var statusStr string
	err := row.Scan(
		&po.ID,
		&po.SupplierID,
		&statusStr,
		&po.TotalAmount,
		&po.CreatedBy,
		&po.ApprovedBy,
		&po.CreatedAt,
		&po.ApprovedAt,
		&po.ReceivedAt,
		&po.Version,
	)
	if err != nil {
		return nil, err
	}
	po.Status = domain.POStatus(statusStr)
	return &po, nil
}

// Compile-time interface assertion.
var _ store.PurchaseOrderRepository = (*PurchaseOrderStore)(nil)

// ---------------------------------------------------------------------------
// POLineStore
// ---------------------------------------------------------------------------

// POLineStore implements store.POLineRepository using a pgx pool.
type POLineStore struct {
	pool *pgxpool.Pool
}

// NewPOLineStore creates a new POLineStore backed by the given connection pool.
func NewPOLineStore(pool *pgxpool.Pool) *POLineStore {
	return &POLineStore{pool: pool}
}

// Create inserts a new purchase order line into the database.
func (s *POLineStore) Create(ctx context.Context, line *domain.PurchaseOrderLine) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		INSERT INTO purchase_order_lines
		    (id, purchase_order_id, item_id, ordered_quantity, ordered_unit_price,
		     received_quantity, received_unit_price)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`

	_, err := db.Exec(ctx, q,
		line.ID,
		line.PurchaseOrderID,
		line.ItemID,
		line.OrderedQuantity,
		line.OrderedUnitPrice,
		line.ReceivedQuantity,
		line.ReceivedUnitPrice,
	)
	if err != nil {
		return fmt.Errorf("po_store.POLineStore.Create: %w", err)
	}
	return nil
}

// GetByID returns a purchase order line by its UUID.
func (s *POLineStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.PurchaseOrderLine, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `
		SELECT id, purchase_order_id, item_id, ordered_quantity, ordered_unit_price,
		       received_quantity, received_unit_price
		FROM purchase_order_lines
		WHERE id = $1
		LIMIT 1`

	line, err := scanPOLine(db.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("po_store.POLineStore.GetByID: %w", err)
	}
	return line, nil
}

// ListByPOID returns all purchase order lines for the given purchase order ID.
func (s *POLineStore) ListByPOID(ctx context.Context, poID uuid.UUID) ([]domain.PurchaseOrderLine, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `
		SELECT id, purchase_order_id, item_id, ordered_quantity, ordered_unit_price,
		       received_quantity, received_unit_price
		FROM purchase_order_lines
		WHERE purchase_order_id = $1
		ORDER BY id`

	rows, err := db.Query(ctx, q, poID)
	if err != nil {
		return nil, fmt.Errorf("po_store.POLineStore.ListByPOID: %w", err)
	}
	defer rows.Close()

	var lines []domain.PurchaseOrderLine
	for rows.Next() {
		line, err := scanPOLine(rows)
		if err != nil {
			return nil, fmt.Errorf("po_store.POLineStore.ListByPOID: %w", err)
		}
		lines = append(lines, *line)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("po_store.POLineStore.ListByPOID: %w", err)
	}
	return lines, nil
}

// Update persists mutable received fields of a purchase order line.
func (s *POLineStore) Update(ctx context.Context, line *domain.PurchaseOrderLine) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		UPDATE purchase_order_lines
		SET received_quantity   = $1,
		    received_unit_price = $2
		WHERE id = $3`

	_, err := db.Exec(ctx, q,
		line.ReceivedQuantity,
		line.ReceivedUnitPrice,
		line.ID,
	)
	if err != nil {
		return fmt.Errorf("po_store.POLineStore.Update: %w", err)
	}
	return nil
}

// scanPOLine reads a single purchase_order_lines row from a pgx row scanner.
func scanPOLine(row pgx.Row) (*domain.PurchaseOrderLine, error) {
	var line domain.PurchaseOrderLine
	err := row.Scan(
		&line.ID,
		&line.PurchaseOrderID,
		&line.ItemID,
		&line.OrderedQuantity,
		&line.OrderedUnitPrice,
		&line.ReceivedQuantity,
		&line.ReceivedUnitPrice,
	)
	if err != nil {
		return nil, err
	}
	return &line, nil
}

// Compile-time interface assertion.
var _ store.POLineRepository = (*POLineStore)(nil)

