package domain

import "fmt"

// ValidateItemForPublish checks whether an item meets all requirements for publishing.
// Returns nil if the item is valid for publishing, or an ErrPublishBlocked with
// accumulated reasons if any checks fail.
func ValidateItemForPublish(item Item, availWindows []AvailabilityWindow, blackoutWindows []BlackoutWindow) *ErrPublishBlocked {
	var reasons []string

	if item.Name == "" {
		reasons = append(reasons, "name is required")
	}
	if item.Category == "" {
		reasons = append(reasons, "category is required")
	}
	if item.Brand == "" {
		reasons = append(reasons, "brand is required")
	}
	if !item.Condition.IsValid() {
		reasons = append(reasons, "item condition is invalid")
	}
	if !item.BillingModel.IsValid() {
		reasons = append(reasons, "billing model is invalid")
	}
	if item.Quantity < 0 {
		reasons = append(reasons, "quantity must be non-negative")
	}

	overlaps := DetectWindowOverlap(availWindows, blackoutWindows)
	reasons = append(reasons, overlaps...)

	if len(reasons) == 0 {
		return nil
	}
	return &ErrPublishBlocked{Reasons: reasons}
}

// DetectWindowOverlap checks for temporal overlaps between availability windows
// and blackout windows. Returns a description string for each overlapping pair found.
func DetectWindowOverlap(availability []AvailabilityWindow, blackouts []BlackoutWindow) []string {
	var overlaps []string

	for _, avail := range availability {
		for _, blackout := range blackouts {
			// Two intervals [a, b) and [c, d) overlap when a < d and c < b
			if avail.StartTime.Before(blackout.EndTime) && blackout.StartTime.Before(avail.EndTime) {
				msg := fmt.Sprintf(
					"availability window [%s, %s) overlaps with blackout window [%s, %s)",
					avail.StartTime.Format("2006-01-02 15:04"),
					avail.EndTime.Format("2006-01-02 15:04"),
					blackout.StartTime.Format("2006-01-02 15:04"),
					blackout.EndTime.Format("2006-01-02 15:04"),
				)
				overlaps = append(overlaps, msg)
			}
		}
	}

	return overlaps
}

// ValidateQuantityNonNegative returns an ErrValidation if the quantity is negative.
func ValidateQuantityNonNegative(qty int) error {
	if qty < 0 {
		return &ErrValidation{
			Field:   "quantity",
			Message: "quantity must be non-negative",
		}
	}
	return nil
}

// ApplyDepositDefault returns the default refundable deposit if the given
// deposit value is zero; otherwise returns the original value.
func ApplyDepositDefault(deposit float64) float64 {
	if deposit == 0 {
		return DefaultRefundableDeposit
	}
	return deposit
}
