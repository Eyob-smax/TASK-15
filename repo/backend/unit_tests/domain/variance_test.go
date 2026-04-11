package domain_test

import (
	"math"
	"testing"
	"time"

	"fitcommerce/internal/domain"

	"github.com/google/uuid"
)

// ─── RequiresEscalation ────────────────────────────────────────────────────────

func TestVarianceRecord_RequiresEscalation_HighAmount(t *testing.T) {
	v := &domain.VarianceRecord{
		ID:               uuid.New(),
		ExpectedValue:    10000.00,
		ActualValue:      9700.00,
		DifferenceAmount: 300.00, // > $250 threshold
		Status:           domain.VarianceStatusOpen,
	}
	if !v.RequiresEscalation() {
		t.Error("expected escalation for difference > $250")
	}
}

func TestVarianceRecord_RequiresEscalation_HighPercent(t *testing.T) {
	v := &domain.VarianceRecord{
		ID:               uuid.New(),
		ExpectedValue:    5000.00,
		ActualValue:      4850.00,
		DifferenceAmount: 150.00, // 150/5000 = 3% > 2% threshold, but $150 < $250
		Status:           domain.VarianceStatusOpen,
	}
	if !v.RequiresEscalation() {
		t.Error("expected escalation for percentage > 2%")
	}
}

func TestVarianceRecord_RequiresEscalation_LowValues(t *testing.T) {
	v := &domain.VarianceRecord{
		ID:               uuid.New(),
		ExpectedValue:    50000.00,
		ActualValue:      49800.00,
		DifferenceAmount: 200.00, // $200 < $250, and 200/50000 = 0.4% < 2%
		Status:           domain.VarianceStatusOpen,
	}
	if v.RequiresEscalation() {
		t.Error("expected no escalation for small difference")
	}
}

func TestVarianceRecord_RequiresEscalation_BothThresholds(t *testing.T) {
	v := &domain.VarianceRecord{
		ID:               uuid.New(),
		ExpectedValue:    1000.00,
		ActualValue:      600.00,
		DifferenceAmount: 400.00, // $400 > $250 AND 40% > 2%
		Status:           domain.VarianceStatusOpen,
	}
	if !v.RequiresEscalation() {
		t.Error("expected escalation when both thresholds exceeded")
	}
}

// ─── IsOverdue ─────────────────────────────────────────────────────────────────

func TestVarianceRecord_IsOverdue_PastDue(t *testing.T) {
	v := &domain.VarianceRecord{
		ID:                uuid.New(),
		Status:            domain.VarianceStatusOpen,
		ResolutionDueDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
	}
	now := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	if !v.IsOverdue(now) {
		t.Error("expected open + past due to be overdue")
	}
}

func TestVarianceRecord_IsOverdue_NotDue(t *testing.T) {
	v := &domain.VarianceRecord{
		ID:                uuid.New(),
		Status:            domain.VarianceStatusOpen,
		ResolutionDueDate: time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
	}
	now := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	if v.IsOverdue(now) {
		t.Error("expected open + future due to NOT be overdue")
	}
}

func TestVarianceRecord_IsOverdue_Resolved(t *testing.T) {
	resolvedAt := time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC)
	v := &domain.VarianceRecord{
		ID:                uuid.New(),
		Status:            domain.VarianceStatusResolved,
		ResolutionDueDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		ResolvedAt:        &resolvedAt,
	}
	now := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	if v.IsOverdue(now) {
		t.Error("expected resolved variance to NOT be overdue even if past due date")
	}
}

// ─── CalculateResolutionDueDate ────────────────────────────────────────────────

func TestCalculateResolutionDueDate_Weekdays(t *testing.T) {
	// Monday 2026-04-06 + 5 business days = Monday 2026-04-13
	monday := time.Date(2026, 4, 6, 10, 0, 0, 0, time.UTC)
	due := domain.CalculateResolutionDueDate(monday)
	expected := time.Date(2026, 4, 13, 10, 0, 0, 0, time.UTC)
	if !due.Equal(expected) {
		t.Errorf("expected due date %v, got %v (weekday: %s)", expected, due, due.Weekday())
	}
}

func TestCalculateResolutionDueDate_OverWeekend(t *testing.T) {
	// Thursday 2026-04-09 + 5 business days:
	// Fri 10, (skip Sat 11, Sun 12), Mon 13, Tue 14, Wed 15, Thu 16
	// That's 5 business days: Fri, Mon, Tue, Wed, Thu = Thu 2026-04-16
	thursday := time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC)
	due := domain.CalculateResolutionDueDate(thursday)
	expected := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)
	if !due.Equal(expected) {
		t.Errorf("expected due date %v, got %v (weekday: %s)", expected, due, due.Weekday())
	}
}

func TestCalculateResolutionDueDate_FridayReceipt(t *testing.T) {
	// Friday 2026-04-10 + 5 business days:
	// (skip Sat 11, Sun 12), Mon 13, Tue 14, Wed 15, Thu 16, Fri 17
	// That's 5 business days: Mon, Tue, Wed, Thu, Fri = Fri 2026-04-17
	friday := time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC)
	due := domain.CalculateResolutionDueDate(friday)
	expected := time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	if !due.Equal(expected) {
		t.Errorf("expected due date %v, got %v (weekday: %s)", expected, due, due.Weekday())
	}
}

// ─── CalculateValueWeightedAllocation ──────────────────────────────────────────

func TestCalculateValueWeightedAllocation_EqualValues(t *testing.T) {
	lines := []domain.PurchaseOrderLine{
		{ID: uuid.New(), OrderedQuantity: 10, OrderedUnitPrice: 100.00},
		{ID: uuid.New(), OrderedQuantity: 10, OrderedUnitPrice: 100.00},
	}
	sharedCost := 500.00
	alloc := domain.CalculateValueWeightedAllocation(lines, sharedCost)

	if len(alloc) != 2 {
		t.Fatalf("expected 2 allocations, got %d", len(alloc))
	}

	for _, line := range lines {
		got := alloc[line.ID]
		if math.Abs(got-250.00) > 0.01 {
			t.Errorf("expected equal split of 250.00, got %f for line %s", got, line.ID)
		}
	}
}

func TestCalculateValueWeightedAllocation_UnequalValues(t *testing.T) {
	line1 := domain.PurchaseOrderLine{ID: uuid.New(), OrderedQuantity: 10, OrderedUnitPrice: 100.00} // value 1000
	line2 := domain.PurchaseOrderLine{ID: uuid.New(), OrderedQuantity: 5, OrderedUnitPrice: 200.00}  // value 1000
	line3 := domain.PurchaseOrderLine{ID: uuid.New(), OrderedQuantity: 20, OrderedUnitPrice: 50.00}  // value 1000
	// Total value = 3000; equal proportions of 1/3 each
	lines := []domain.PurchaseOrderLine{line1, line2, line3}
	sharedCost := 300.00
	alloc := domain.CalculateValueWeightedAllocation(lines, sharedCost)

	for _, line := range lines {
		got := alloc[line.ID]
		if math.Abs(got-100.00) > 0.01 {
			t.Errorf("expected 100.00 for equal-value line, got %f", got)
		}
	}

	// Now test with truly unequal values
	lineA := domain.PurchaseOrderLine{ID: uuid.New(), OrderedQuantity: 10, OrderedUnitPrice: 10.00} // value 100
	lineB := domain.PurchaseOrderLine{ID: uuid.New(), OrderedQuantity: 10, OrderedUnitPrice: 30.00} // value 300
	// Total value = 400; A gets 25%, B gets 75%
	lines2 := []domain.PurchaseOrderLine{lineA, lineB}
	alloc2 := domain.CalculateValueWeightedAllocation(lines2, 400.00)
	if math.Abs(alloc2[lineA.ID]-100.00) > 0.01 {
		t.Errorf("expected 100.00 for line A, got %f", alloc2[lineA.ID])
	}
	if math.Abs(alloc2[lineB.ID]-300.00) > 0.01 {
		t.Errorf("expected 300.00 for line B, got %f", alloc2[lineB.ID])
	}
}

func TestCalculateValueWeightedAllocation_SingleLine(t *testing.T) {
	line := domain.PurchaseOrderLine{ID: uuid.New(), OrderedQuantity: 5, OrderedUnitPrice: 100.00}
	lines := []domain.PurchaseOrderLine{line}
	sharedCost := 750.00
	alloc := domain.CalculateValueWeightedAllocation(lines, sharedCost)
	if math.Abs(alloc[line.ID]-750.00) > 0.01 {
		t.Errorf("expected single line to get full 750.00, got %f", alloc[line.ID])
	}
}

func TestCalculateValueWeightedAllocation_ZeroTotal(t *testing.T) {
	// Lines with zero value get even distribution
	line1 := domain.PurchaseOrderLine{ID: uuid.New(), OrderedQuantity: 0, OrderedUnitPrice: 100.00}
	line2 := domain.PurchaseOrderLine{ID: uuid.New(), OrderedQuantity: 0, OrderedUnitPrice: 50.00}
	lines := []domain.PurchaseOrderLine{line1, line2}
	sharedCost := 200.00
	alloc := domain.CalculateValueWeightedAllocation(lines, sharedCost)

	for _, line := range lines {
		got := alloc[line.ID]
		if math.Abs(got-100.00) > 0.01 {
			t.Errorf("expected even split of 100.00 for zero-value lines, got %f", got)
		}
	}
}
