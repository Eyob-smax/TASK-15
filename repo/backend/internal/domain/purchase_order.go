package domain

import (
	"time"

	"github.com/google/uuid"
)

// PurchaseOrder represents an order placed with a supplier for procurement.
type PurchaseOrder struct {
	ID          uuid.UUID
	SupplierID  uuid.UUID
	Status      POStatus
	TotalAmount float64
	CreatedBy   uuid.UUID
	ApprovedBy  *uuid.UUID
	CreatedAt   time.Time
	ApprovedAt  *time.Time
	ReceivedAt  *time.Time
	Version     int
}

// PurchaseOrderLine represents a single line item within a purchase order.
type PurchaseOrderLine struct {
	ID                uuid.UUID
	PurchaseOrderID   uuid.UUID
	ItemID            uuid.UUID
	OrderedQuantity   int
	OrderedUnitPrice  float64
	ReceivedQuantity  *int
	ReceivedUnitPrice *float64
}
