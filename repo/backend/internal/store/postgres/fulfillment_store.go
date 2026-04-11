package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
)

// FulfillmentStore implements store.FulfillmentRepository using PostgreSQL.
// It persists fulfillment groups and order-to-group associations created during
// order split and merge operations.
type FulfillmentStore struct {
	pool *pgxpool.Pool
}

// NewFulfillmentStore creates a FulfillmentStore backed by the given pool.
func NewFulfillmentStore(pool *pgxpool.Pool) *FulfillmentStore {
	return &FulfillmentStore{pool: pool}
}

// CreateGroup inserts a new fulfillment group record.
func (s *FulfillmentStore) CreateGroup(ctx context.Context, group *domain.FulfillmentGroup) error {
	db := executorFromContext(ctx, s.pool)
	_, err := db.Exec(ctx,
		`INSERT INTO fulfillment_groups
		 (id, supplier_id, warehouse_bin_id, pickup_point, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		group.ID, group.SupplierID, group.WarehouseBinID,
		group.PickupPoint, group.Status, group.CreatedAt,
	)
	return err
}

// AddGroupOrder inserts a fulfillment_group_orders row linking an order to a group.
func (s *FulfillmentStore) AddGroupOrder(ctx context.Context, entry *domain.FulfillmentGroupOrder) error {
	db := executorFromContext(ctx, s.pool)
	_, err := db.Exec(ctx,
		`INSERT INTO fulfillment_group_orders
		 (id, fulfillment_group_id, order_id, quantity)
		 VALUES ($1, $2, $3, $4)`,
		entry.ID, entry.FulfillmentGroupID, entry.OrderID, entry.Quantity,
	)
	return err
}
