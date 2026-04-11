package domain

import (
	"time"

	"github.com/google/uuid"
)

// FulfillmentGroup associates a set of split or merged orders with a specific
// supplier, warehouse bin, and optional pickup point for end-to-end traceability.
type FulfillmentGroup struct {
	ID             uuid.UUID
	SupplierID     *uuid.UUID
	WarehouseBinID *uuid.UUID
	PickupPoint    string
	Status         string
	CreatedAt      time.Time
}

// FulfillmentGroupOrder links a child order to its fulfillment group with the
// quantity allocated to that group.
type FulfillmentGroupOrder struct {
	ID                 uuid.UUID
	FulfillmentGroupID uuid.UUID
	OrderID            uuid.UUID
	Quantity           int
}
