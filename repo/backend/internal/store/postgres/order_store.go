package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
)

// Compile-time interface compliance assertions.
var (
	_ store.OrderRepository    = (*OrderStore)(nil)
	_ store.TimelineRepository = (*TimelineStore)(nil)
)

// OrderStore implements OrderRepository.
type OrderStore struct {
	pool *pgxpool.Pool
}

// NewOrderStore creates a new OrderStore backed by the given pool.
func NewOrderStore(pool *pgxpool.Pool) *OrderStore {
	return &OrderStore{pool: pool}
}

// TimelineStore implements TimelineRepository.
type TimelineStore struct {
	pool *pgxpool.Pool
}

// NewTimelineStore creates a new TimelineStore backed by the given pool.
func NewTimelineStore(pool *pgxpool.Pool) *TimelineStore {
	return &TimelineStore{pool: pool}
}

// --- OrderRepository ---

func (s *OrderStore) Create(ctx context.Context, order *domain.Order) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		INSERT INTO orders
			(id, user_id, item_id, campaign_id, quantity, unit_price, total_amount,
			 status, settlement_marker, notes, auto_close_at, created_at, updated_at,
			 paid_at, cancelled_at, refunded_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`
	_, err := db.Exec(ctx, q,
		order.ID, order.UserID, order.ItemID, order.CampaignID,
		order.Quantity, order.UnitPrice, order.TotalAmount,
		string(order.Status), order.SettlementMarker, order.Notes,
		order.AutoCloseAt, order.CreatedAt, order.UpdatedAt,
		order.PaidAt, order.CancelledAt, order.RefundedAt,
	)
	return err
}

func (s *OrderStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `
		SELECT id, user_id, item_id, campaign_id, quantity, unit_price, total_amount,
		       status, settlement_marker, notes, auto_close_at, created_at, updated_at,
		       paid_at, cancelled_at, refunded_at
		FROM orders WHERE id = $1`
	order, err := scanOrder(db.QueryRow(ctx, q, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return order, nil
}

func (s *OrderStore) List(ctx context.Context, userID *uuid.UUID, page, pageSize int) ([]domain.Order, int, error) {
	db := executorFromContext(ctx, s.pool)
	countQ := `SELECT COUNT(*) FROM orders WHERE 1=1`
	listQ := `
		SELECT id, user_id, item_id, campaign_id, quantity, unit_price, total_amount,
		       status, settlement_marker, notes, auto_close_at, created_at, updated_at,
		       paid_at, cancelled_at, refunded_at
		FROM orders WHERE 1=1`
	args := []interface{}{}
	idx := 1

	if userID != nil {
		filter := fmt.Sprintf(` AND user_id = $%d`, idx)
		countQ += filter
		listQ += filter
		args = append(args, *userID)
		idx++
	}

	var total int
	if err := db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQ += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	listArgs := append(args, pageSize, (page-1)*pageSize)

	rows, err := db.Query(ctx, listQ, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, *o)
	}
	return orders, total, rows.Err()
}

func (s *OrderStore) Update(ctx context.Context, order *domain.Order) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		UPDATE orders SET
			status=$1, settlement_marker=$2, notes=$3, updated_at=$4,
			paid_at=$5, cancelled_at=$6, refunded_at=$7
		WHERE id=$8`
	_, err := db.Exec(ctx, q,
		string(order.Status), order.SettlementMarker, order.Notes, order.UpdatedAt,
		order.PaidAt, order.CancelledAt, order.RefundedAt,
		order.ID,
	)
	return err
}

// ListExpiredUnpaid returns Created orders whose auto_close_at is at or before now.
// Used by AutoCloseJob to find orders requiring closure.
func (s *OrderStore) ListExpiredUnpaid(ctx context.Context, now time.Time) ([]domain.Order, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `
		SELECT id, user_id, item_id, campaign_id, quantity, unit_price, total_amount,
		       status, settlement_marker, notes, auto_close_at, created_at, updated_at,
		       paid_at, cancelled_at, refunded_at
		FROM orders
		WHERE status = 'created' AND auto_close_at <= $1
		ORDER BY auto_close_at`
	rows, err := db.Query(ctx, q, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *o)
	}
	return orders, rows.Err()
}

// --- TimelineRepository ---

func (s *TimelineStore) Create(ctx context.Context, entry *domain.OrderTimelineEntry) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		INSERT INTO order_timeline_entries (id, order_id, action, description, performed_by, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := db.Exec(ctx, q,
		entry.ID, entry.OrderID, entry.Action, entry.Description,
		entry.PerformedBy, entry.CreatedAt,
	)
	return err
}

func (s *TimelineStore) ListByOrderID(ctx context.Context, orderID uuid.UUID) ([]domain.OrderTimelineEntry, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `
		SELECT id, order_id, action, description, performed_by, created_at
		FROM order_timeline_entries WHERE order_id = $1 ORDER BY created_at`
	rows, err := db.Query(ctx, q, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.OrderTimelineEntry
	for rows.Next() {
		var e domain.OrderTimelineEntry
		if err := rows.Scan(&e.ID, &e.OrderID, &e.Action, &e.Description, &e.PerformedBy, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// --- Private helpers ---

type orderScannable interface {
	Scan(dest ...interface{}) error
}

func scanOrder(row orderScannable) (*domain.Order, error) {
	var o domain.Order
	var status string
	err := row.Scan(
		&o.ID, &o.UserID, &o.ItemID, &o.CampaignID,
		&o.Quantity, &o.UnitPrice, &o.TotalAmount,
		&status, &o.SettlementMarker, &o.Notes,
		&o.AutoCloseAt, &o.CreatedAt, &o.UpdatedAt,
		&o.PaidAt, &o.CancelledAt, &o.RefundedAt,
	)
	if err != nil {
		return nil, err
	}
	o.Status = domain.OrderStatus(status)
	return &o, nil
}
