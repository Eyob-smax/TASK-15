package domain

import (
	"time"

	"github.com/google/uuid"
)

// InventorySnapshot records the quantity of an item at a specific point in time.
type InventorySnapshot struct {
	ID         uuid.UUID
	ItemID     uuid.UUID
	Quantity   int
	LocationID *uuid.UUID
	RecordedAt time.Time
}

// InventoryAdjustment records a manual change to inventory quantities.
type InventoryAdjustment struct {
	ID             uuid.UUID
	ItemID         uuid.UUID
	QuantityChange int
	Reason         string
	CreatedBy      uuid.UUID
	CreatedAt      time.Time
}

// WarehouseBin represents a physical storage bin within a location's warehouse.
type WarehouseBin struct {
	ID          uuid.UUID
	LocationID  uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
