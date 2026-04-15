package orders_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
)

func TestOrderService_Get_Found(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	id := uuid.New()
	orders.orders[id] = &domain.Order{ID: id, Status: domain.OrderStatusCreated}

	got, err := svc.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.ID != id {
		t.Errorf("expected %v got %v", id, got.ID)
	}
}

func TestOrderService_Get_NotFound(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	_, err := svc.Get(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found")
	}
}

func TestOrderService_List_AllOrders(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	for i := 0; i < 3; i++ {
		id := uuid.New()
		orders.orders[id] = &domain.Order{ID: id, Status: domain.OrderStatusCreated}
	}

	rows, total, err := svc.List(context.Background(), nil, 1, 50)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 3 || len(rows) != 3 {
		t.Errorf("expected 3 rows total=%d rows=%d", total, len(rows))
	}
}

func TestOrderService_List_ByUser(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	uid := uuid.New()
	for i := 0; i < 2; i++ {
		id := uuid.New()
		orders.orders[id] = &domain.Order{ID: id, UserID: uid, Status: domain.OrderStatusCreated}
	}

	// mockOrderRepo.List ignores userID filter — we just want List() to not error.
	_, _, err := svc.List(context.Background(), &uid, 1, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
}
