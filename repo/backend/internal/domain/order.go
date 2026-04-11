package domain

import (
	"time"

	"github.com/google/uuid"
)

// AutoCloseTimeout is the duration after which an unpaid order is automatically closed.
const AutoCloseTimeout = 30 * time.Minute

// Order represents a purchase order placed by a user for an item.
type Order struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	ItemID           uuid.UUID
	CampaignID       *uuid.UUID
	Quantity         int
	UnitPrice        float64
	TotalAmount      float64
	Status           OrderStatus
	SettlementMarker string
	Notes            string
	AutoCloseAt      time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	PaidAt           *time.Time
	CancelledAt      *time.Time
	RefundedAt       *time.Time
}

// ShouldAutoClose returns true if the order is in "created" status and the
// current time is at or past the auto-close deadline.
func (o *Order) ShouldAutoClose(now time.Time) bool {
	return o.Status == OrderStatusCreated && !now.Before(o.AutoCloseAt)
}

// OrderTimelineEntry records a chronological event in an order's lifecycle.
type OrderTimelineEntry struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	Action      string
	Description string
	PerformedBy uuid.UUID
	CreatedAt   time.Time
}
