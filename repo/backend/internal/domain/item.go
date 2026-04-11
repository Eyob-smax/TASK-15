package domain

import (
	"time"

	"github.com/google/uuid"
)

// DefaultRefundableDeposit is the default deposit amount for items when none is specified.
const DefaultRefundableDeposit = 50.00

// Item represents a piece of equipment or product in the catalog.
type Item struct {
	ID                uuid.UUID
	SKU               string
	Name              string
	Description       string
	Category          string
	Brand             string
	Condition         ItemCondition
	UnitPrice         float64
	RefundableDeposit float64
	BillingModel      BillingModel
	Status            ItemStatus
	Quantity          int
	LocationID        *uuid.UUID
	CreatedBy         uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Version           int
}

// ApplyDepositDefault sets the refundable deposit to the default value if it is zero.
func (i *Item) ApplyDepositDefault() {
	if i.RefundableDeposit == 0 {
		i.RefundableDeposit = DefaultRefundableDeposit
	}
}

// Validate checks all required fields and constraints on the item.
// Returns a slice of ValidationError for each issue found, or nil if valid.
func (i *Item) Validate() []ValidationError {
	var errs []ValidationError

	if i.Name == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "name is required"})
	}
	if i.Category == "" {
		errs = append(errs, ValidationError{Field: "category", Message: "category is required"})
	}
	if i.Brand == "" {
		errs = append(errs, ValidationError{Field: "brand", Message: "brand is required"})
	}
	if !i.Condition.IsValid() {
		errs = append(errs, ValidationError{Field: "condition", Message: "invalid item condition"})
	}
	if !i.BillingModel.IsValid() {
		errs = append(errs, ValidationError{Field: "billing_model", Message: "invalid billing model"})
	}
	if i.Quantity < 0 {
		errs = append(errs, ValidationError{Field: "quantity", Message: "quantity must be non-negative"})
	}
	if i.RefundableDeposit < 0 {
		errs = append(errs, ValidationError{Field: "refundable_deposit", Message: "deposit must be non-negative"})
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

// AvailabilityWindow represents a time window during which an item is available.
type AvailabilityWindow struct {
	ID        uuid.UUID
	ItemID    uuid.UUID
	StartTime time.Time
	EndTime   time.Time
}

// BlackoutWindow represents a time window during which an item is not available.
type BlackoutWindow struct {
	ID        uuid.UUID
	ItemID    uuid.UUID
	StartTime time.Time
	EndTime   time.Time
}
