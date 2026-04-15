package procurement_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
)

func TestPurchaseOrderService_List_ReturnsAll(t *testing.T) {
	poRepo := newMockPORepo()
	svc := newPOService(poRepo, newMockPOLineRepo(), newMockVarianceRepo(),
		newMockLandedCostRepo(), newMockInventoryRepo(), newMockItemRepo())

	for i := 0; i < 3; i++ {
		id := uuid.New()
		poRepo.orders[id] = &domain.PurchaseOrder{ID: id, SupplierID: uuid.New(), Status: domain.POStatusCreated}
	}

	rows, total, err := svc.List(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 3 || len(rows) != 3 {
		t.Errorf("expected 3, got total=%d rows=%d", total, len(rows))
	}
}

func TestPurchaseOrderService_List_Empty(t *testing.T) {
	poRepo := newMockPORepo()
	svc := newPOService(poRepo, newMockPOLineRepo(), newMockVarianceRepo(),
		newMockLandedCostRepo(), newMockInventoryRepo(), newMockItemRepo())

	rows, total, err := svc.List(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 0 || len(rows) != 0 {
		t.Errorf("expected empty, got total=%d rows=%d", total, len(rows))
	}
}
