package domain

import (
	"math"
	"time"

	"github.com/google/uuid"
)

// Variance escalation thresholds.
const (
	VarianceEscalationAmountThreshold  = 250.00
	VarianceEscalationPercentThreshold = 0.02 // 2%
	VarianceResolutionBusinessDays     = 5
)

// VarianceRecord tracks discrepancies between expected and actual values
// during purchase order receipt.
type VarianceRecord struct {
	ID                uuid.UUID
	POLineID          uuid.UUID
	Type              VarianceType
	ExpectedValue     float64
	ActualValue       float64
	DifferenceAmount  float64
	Status            VarianceStatus
	ResolutionDueDate time.Time
	ResolvedAt        *time.Time
	ResolutionAction  string
	ResolutionNotes   string
	QuantityChange    *int
	CreatedAt         time.Time
}

// RequiresEscalation returns true if the absolute difference exceeds the amount
// threshold or the percentage threshold relative to the expected value.
func (v *VarianceRecord) RequiresEscalation() bool {
	absDiff := math.Abs(v.DifferenceAmount)
	if absDiff > VarianceEscalationAmountThreshold {
		return true
	}
	if v.ExpectedValue != 0 && math.Abs(v.DifferenceAmount/v.ExpectedValue) > VarianceEscalationPercentThreshold {
		return true
	}
	return false
}

// IsOverdue returns true if the variance is still open and the current time
// is past the resolution due date.
func (v *VarianceRecord) IsOverdue(now time.Time) bool {
	return v.Status == VarianceStatusOpen && now.After(v.ResolutionDueDate)
}

// CalculateResolutionDueDate computes the resolution due date by adding
// VarianceResolutionBusinessDays business days (skipping weekends) to the receipt date.
func CalculateResolutionDueDate(receiptDate time.Time) time.Time {
	dueDate := receiptDate
	businessDaysAdded := 0
	for businessDaysAdded < VarianceResolutionBusinessDays {
		dueDate = dueDate.Add(24 * time.Hour)
		weekday := dueDate.Weekday()
		if weekday != time.Saturday && weekday != time.Sunday {
			businessDaysAdded++
		}
	}
	return dueDate
}

// CalculateValueWeightedAllocation distributes a shared cost across purchase order
// lines proportionally by each line's total value (ordered quantity * ordered unit price).
// Returns a map from PurchaseOrderLine ID to allocated amount.
func CalculateValueWeightedAllocation(lines []PurchaseOrderLine, sharedCost float64) map[uuid.UUID]float64 {
	allocations := make(map[uuid.UUID]float64, len(lines))
	if len(lines) == 0 {
		return allocations
	}

	totalValue := 0.0
	lineValues := make(map[uuid.UUID]float64, len(lines))
	for _, line := range lines {
		value := float64(line.OrderedQuantity) * line.OrderedUnitPrice
		lineValues[line.ID] = value
		totalValue += value
	}

	if totalValue == 0 {
		// If total value is zero, distribute evenly
		evenShare := sharedCost / float64(len(lines))
		for _, line := range lines {
			allocations[line.ID] = evenShare
		}
		return allocations
	}

	for _, line := range lines {
		proportion := lineValues[line.ID] / totalValue
		allocations[line.ID] = sharedCost * proportion
	}
	return allocations
}

// LandedCostEntry records a cost component allocated to a specific item
// from a purchase order for a given accounting period.
type LandedCostEntry struct {
	ID               uuid.UUID
	ItemID           uuid.UUID
	PurchaseOrderID  uuid.UUID
	POLineID         uuid.UUID
	Period           string // e.g., "2026-Q1"
	CostComponent    string // e.g., "unit_cost", "freight", "duty"
	RawAmount        float64
	AllocatedAmount  float64
	AllocationMethod string // "value_weighted" or "direct"
	CreatedAt        time.Time
}
