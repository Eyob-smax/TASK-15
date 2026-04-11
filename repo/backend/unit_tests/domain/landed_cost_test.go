package domain_test

import (
	"math"
	"testing"

	"fitcommerce/internal/domain"

	"github.com/google/uuid"
)

func TestCalculateValueWeightedAllocation_ThreeLines(t *testing.T) {
	line1 := domain.PurchaseOrderLine{ID: uuid.New(), OrderedQuantity: 10, OrderedUnitPrice: 50.00}  // value 500
	line2 := domain.PurchaseOrderLine{ID: uuid.New(), OrderedQuantity: 5, OrderedUnitPrice: 200.00}  // value 1000
	line3 := domain.PurchaseOrderLine{ID: uuid.New(), OrderedQuantity: 20, OrderedUnitPrice: 25.00}  // value 500
	// Total value = 2000
	// line1: 500/2000 = 25%
	// line2: 1000/2000 = 50%
	// line3: 500/2000 = 25%
	lines := []domain.PurchaseOrderLine{line1, line2, line3}
	sharedCost := 1000.00

	alloc := domain.CalculateValueWeightedAllocation(lines, sharedCost)

	if len(alloc) != 3 {
		t.Fatalf("expected 3 allocations, got %d", len(alloc))
	}

	expected1 := 250.00 // 25% of 1000
	expected2 := 500.00 // 50% of 1000
	expected3 := 250.00 // 25% of 1000

	if math.Abs(alloc[line1.ID]-expected1) > 0.01 {
		t.Errorf("line1: expected %.2f, got %.2f", expected1, alloc[line1.ID])
	}
	if math.Abs(alloc[line2.ID]-expected2) > 0.01 {
		t.Errorf("line2: expected %.2f, got %.2f", expected2, alloc[line2.ID])
	}
	if math.Abs(alloc[line3.ID]-expected3) > 0.01 {
		t.Errorf("line3: expected %.2f, got %.2f", expected3, alloc[line3.ID])
	}

	// Verify sum equals shared cost
	total := alloc[line1.ID] + alloc[line2.ID] + alloc[line3.ID]
	if math.Abs(total-sharedCost) > 0.01 {
		t.Errorf("sum of allocations %.2f does not equal shared cost %.2f", total, sharedCost)
	}
}

func TestCalculateValueWeightedAllocation_Precision(t *testing.T) {
	// Use values that could cause floating-point drift:
	// 3 lines with values that create repeating decimals
	line1 := domain.PurchaseOrderLine{ID: uuid.New(), OrderedQuantity: 1, OrderedUnitPrice: 33.33}
	line2 := domain.PurchaseOrderLine{ID: uuid.New(), OrderedQuantity: 1, OrderedUnitPrice: 33.33}
	line3 := domain.PurchaseOrderLine{ID: uuid.New(), OrderedQuantity: 1, OrderedUnitPrice: 33.34}
	lines := []domain.PurchaseOrderLine{line1, line2, line3}
	sharedCost := 100.00

	alloc := domain.CalculateValueWeightedAllocation(lines, sharedCost)

	total := 0.0
	for _, amount := range alloc {
		total += amount
	}

	epsilon := 0.01
	if math.Abs(total-sharedCost) > epsilon {
		t.Errorf("sum of allocations %.6f differs from shared cost %.2f by more than %.2f", total, sharedCost, epsilon)
	}
}

func TestCalculateValueWeightedAllocation_ZeroValueLine(t *testing.T) {
	// One line with zero value alongside non-zero lines
	line1 := domain.PurchaseOrderLine{ID: uuid.New(), OrderedQuantity: 0, OrderedUnitPrice: 100.00} // value 0
	line2 := domain.PurchaseOrderLine{ID: uuid.New(), OrderedQuantity: 10, OrderedUnitPrice: 50.00} // value 500
	lines := []domain.PurchaseOrderLine{line1, line2}
	sharedCost := 200.00

	alloc := domain.CalculateValueWeightedAllocation(lines, sharedCost)

	// line1 has 0 value, so gets 0 allocation
	if math.Abs(alloc[line1.ID]-0.00) > 0.01 {
		t.Errorf("expected 0.00 for zero-value line, got %.2f", alloc[line1.ID])
	}
	// line2 gets entire shared cost
	if math.Abs(alloc[line2.ID]-200.00) > 0.01 {
		t.Errorf("expected 200.00 for sole non-zero line, got %.2f", alloc[line2.ID])
	}
}
