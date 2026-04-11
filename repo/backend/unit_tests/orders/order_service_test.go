package orders_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
)

// --- Mock repositories ---

type mockOrderRepo struct {
	orders map[uuid.UUID]*domain.Order
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{orders: make(map[uuid.UUID]*domain.Order)}
}

func (m *mockOrderRepo) Create(_ context.Context, o *domain.Order) error {
	m.orders[o.ID] = o
	return nil
}
func (m *mockOrderRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Order, error) {
	o, ok := m.orders[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return o, nil
}
func (m *mockOrderRepo) List(_ context.Context, _ *uuid.UUID, _, _ int) ([]domain.Order, int, error) {
	list := make([]domain.Order, 0, len(m.orders))
	for _, o := range m.orders {
		list = append(list, *o)
	}
	return list, len(list), nil
}
func (m *mockOrderRepo) Update(_ context.Context, o *domain.Order) error {
	m.orders[o.ID] = o
	return nil
}
func (m *mockOrderRepo) ListExpiredUnpaid(_ context.Context, now time.Time) ([]domain.Order, error) {
	var result []domain.Order
	for _, o := range m.orders {
		if o.Status == domain.OrderStatusCreated && !now.Before(o.AutoCloseAt) {
			result = append(result, *o)
		}
	}
	return result, nil
}

type mockTimelineRepo struct {
	entries []domain.OrderTimelineEntry
}

func newMockTimelineRepo() *mockTimelineRepo { return &mockTimelineRepo{} }

func (m *mockTimelineRepo) Create(_ context.Context, e *domain.OrderTimelineEntry) error {
	m.entries = append(m.entries, *e)
	return nil
}
func (m *mockTimelineRepo) ListByOrderID(_ context.Context, orderID uuid.UUID) ([]domain.OrderTimelineEntry, error) {
	var result []domain.OrderTimelineEntry
	for _, e := range m.entries {
		if e.OrderID == orderID {
			result = append(result, e)
		}
	}
	return result, nil
}

type mockItemRepo struct {
	items map[uuid.UUID]*domain.Item
}

func newMockItemRepo() *mockItemRepo {
	return &mockItemRepo{items: make(map[uuid.UUID]*domain.Item)}
}

func (m *mockItemRepo) Create(_ context.Context, item *domain.Item) error {
	m.items[item.ID] = item
	return nil
}
func (m *mockItemRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Item, error) {
	item, ok := m.items[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return item, nil
}
func (m *mockItemRepo) List(_ context.Context, _ map[string]string, _, _ int) ([]domain.Item, int, error) {
	return nil, 0, nil
}
func (m *mockItemRepo) Update(_ context.Context, item *domain.Item) error {
	m.items[item.ID] = item
	return nil
}
func (m *mockItemRepo) BatchUpdate(_ context.Context, _ []*domain.Item) error { return nil }

type mockInventoryRepo struct {
	adjustments []domain.InventoryAdjustment
	snapshots   []domain.InventorySnapshot
}

func newMockInventoryRepo() *mockInventoryRepo { return &mockInventoryRepo{} }

func (m *mockInventoryRepo) CreateSnapshot(_ context.Context, snapshot *domain.InventorySnapshot) error {
	m.snapshots = append(m.snapshots, *snapshot)
	return nil
}
func (m *mockInventoryRepo) ListSnapshots(_ context.Context, _ *uuid.UUID, _ *uuid.UUID) ([]domain.InventorySnapshot, error) {
	return nil, nil
}
func (m *mockInventoryRepo) CreateAdjustment(_ context.Context, adj *domain.InventoryAdjustment) error {
	m.adjustments = append(m.adjustments, *adj)
	return nil
}
func (m *mockInventoryRepo) ListAdjustments(_ context.Context, _ *uuid.UUID, _, _ int) ([]domain.InventoryAdjustment, int, error) {
	return nil, 0, nil
}

type mockAvailabilityRepo struct {
	windows map[uuid.UUID][]domain.AvailabilityWindow
}

func newMockAvailabilityRepo() *mockAvailabilityRepo {
	return &mockAvailabilityRepo{windows: make(map[uuid.UUID][]domain.AvailabilityWindow)}
}

func (m *mockAvailabilityRepo) Create(_ context.Context, window *domain.AvailabilityWindow) error {
	m.windows[window.ItemID] = append(m.windows[window.ItemID], *window)
	return nil
}
func (m *mockAvailabilityRepo) ListByItemID(_ context.Context, itemID uuid.UUID) ([]domain.AvailabilityWindow, error) {
	return append([]domain.AvailabilityWindow(nil), m.windows[itemID]...), nil
}
func (m *mockAvailabilityRepo) DeleteByItemID(_ context.Context, itemID uuid.UUID) error {
	delete(m.windows, itemID)
	return nil
}

type mockBlackoutRepo struct {
	windows map[uuid.UUID][]domain.BlackoutWindow
}

func newMockBlackoutRepo() *mockBlackoutRepo {
	return &mockBlackoutRepo{windows: make(map[uuid.UUID][]domain.BlackoutWindow)}
}

func (m *mockBlackoutRepo) Create(_ context.Context, window *domain.BlackoutWindow) error {
	m.windows[window.ItemID] = append(m.windows[window.ItemID], *window)
	return nil
}
func (m *mockBlackoutRepo) ListByItemID(_ context.Context, itemID uuid.UUID) ([]domain.BlackoutWindow, error) {
	return append([]domain.BlackoutWindow(nil), m.windows[itemID]...), nil
}
func (m *mockBlackoutRepo) DeleteByItemID(_ context.Context, itemID uuid.UUID) error {
	delete(m.windows, itemID)
	return nil
}

type mockAuditService struct{}

func (m *mockAuditService) Log(_ context.Context, _, _ string, _, _ uuid.UUID, _ map[string]interface{}) error {
	return nil
}
func (m *mockAuditService) List(_ context.Context, _ string, _ *uuid.UUID, _, _ int) ([]domain.AuditEvent, int, error) {
	return nil, 0, nil
}
func (m *mockAuditService) ListByEventTypes(_ context.Context, _ []string, _, _ int) ([]domain.AuditEvent, int, error) {
	return nil, 0, nil
}

type mockFulfillmentRepo struct{}

func (m *mockFulfillmentRepo) CreateGroup(_ context.Context, _ *domain.FulfillmentGroup) error {
	return nil
}
func (m *mockFulfillmentRepo) AddGroupOrder(_ context.Context, _ *domain.FulfillmentGroupOrder) error {
	return nil
}

func newOrderService(orders *mockOrderRepo, timelines *mockTimelineRepo, items *mockItemRepo, inventory *mockInventoryRepo) *application.OrderServiceImpl {
	return application.NewOrderService(
		orders,
		timelines,
		items,
		inventory,
		newMockAvailabilityRepo(),
		newMockBlackoutRepo(),
		&mockFulfillmentRepo{},
		&mockAuditService{},
		nil,
	)
}

// --- Tests ---

func TestOrderCreate_NonPublishedItem_ReturnsValidation(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	itemID := uuid.New()
	items.items[itemID] = &domain.Item{ID: itemID, Status: domain.ItemStatusDraft}

	o := &domain.Order{UserID: uuid.New(), ItemID: itemID, Quantity: 1}
	_, err := svc.Create(context.Background(), o)
	if err == nil {
		t.Fatal("expected validation error for non-published item")
	}
}

func TestOrderCreate_Success_SetsInventoryAndAutoClose(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	itemID := uuid.New()
	items.items[itemID] = &domain.Item{
		ID:                itemID,
		Status:            domain.ItemStatusPublished,
		UnitPrice:         130,
		RefundableDeposit: 100,
		Quantity:          10,
		Version:           1,
	}

	o := &domain.Order{UserID: uuid.New(), ItemID: itemID, Quantity: 3}
	created, err := svc.Create(context.Background(), o)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// AutoCloseAt should be ~30 minutes from now.
	expected := time.Now().Add(domain.AutoCloseTimeout)
	diff := created.AutoCloseAt.Sub(expected)
	if diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("AutoCloseAt not ~30min from now: %v", created.AutoCloseAt)
	}

	// Inventory adjustment of -3.
	if len(inventory.adjustments) != 1 || inventory.adjustments[0].QuantityChange != -3 {
		t.Error("expected inventory adjustment of -3")
	}
	if created.UnitPrice != 130 {
		t.Errorf("expected unit price from item.unit_price (130), got %v", created.UnitPrice)
	}
	if created.TotalAmount != 390 {
		t.Errorf("expected total amount 390, got %v", created.TotalAmount)
	}
	if len(inventory.snapshots) != 1 || inventory.snapshots[0].Quantity != 7 {
		t.Error("expected inventory snapshot with post-reservation quantity 7")
	}
}

func TestOrderCreate_InsufficientStock_ReturnsValidation(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	itemID := uuid.New()
	items.items[itemID] = &domain.Item{
		ID:                itemID,
		Status:            domain.ItemStatusPublished,
		RefundableDeposit: 25,
		Quantity:          2,
		Version:           1,
	}

	_, err := svc.Create(context.Background(), &domain.Order{UserID: uuid.New(), ItemID: itemID, Quantity: 3})
	if err == nil {
		t.Fatal("expected validation error for insufficient stock")
	}
}

func TestOrderCreate_BlackoutWindow_ReturnsValidation(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	availability := newMockAvailabilityRepo()
	blackouts := newMockBlackoutRepo()
	svc := application.NewOrderService(orders, timelines, items, inventory, availability, blackouts, &mockFulfillmentRepo{}, &mockAuditService{}, nil)

	itemID := uuid.New()
	items.items[itemID] = &domain.Item{
		ID:                itemID,
		Status:            domain.ItemStatusPublished,
		RefundableDeposit: 25,
		Quantity:          5,
		Version:           1,
	}
	blackouts.windows[itemID] = []domain.BlackoutWindow{{
		ID:        uuid.New(),
		ItemID:    itemID,
		StartTime: time.Now().Add(-1 * time.Hour),
		EndTime:   time.Now().Add(1 * time.Hour),
	}}

	_, err := svc.Create(context.Background(), &domain.Order{UserID: uuid.New(), ItemID: itemID, Quantity: 1})
	if err == nil {
		t.Fatal("expected validation error for blackout window")
	}
}

func TestOrderPay_CreatedToPaid_Success(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	id := uuid.New()
	orders.orders[id] = &domain.Order{ID: id, UserID: uuid.New(), Status: domain.OrderStatusCreated}

	performedBy := uuid.New()
	if err := svc.Pay(context.Background(), id, "settlement-abc", performedBy); err != nil {
		t.Fatalf("Pay failed: %v", err)
	}
	if orders.orders[id].Status != domain.OrderStatusPaid {
		t.Errorf("expected status=paid, got %v", orders.orders[id].Status)
	}
	if len(timelines.entries) == 0 || timelines.entries[len(timelines.entries)-1].PerformedBy != performedBy {
		t.Fatalf("expected paid timeline entry performed by %s", performedBy)
	}
}

func TestOrderPay_CancelledOrder_ReturnsInvalidTransition(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	id := uuid.New()
	orders.orders[id] = &domain.Order{ID: id, UserID: uuid.New(), Status: domain.OrderStatusCancelled}

	if err := svc.Pay(context.Background(), id, "x", uuid.New()); err == nil {
		t.Fatal("expected invalid transition error")
	}
}

func TestOrderCancel_CreatedOrder_RestoresInventory(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	id := uuid.New()
	itemID := uuid.New()
	items.items[itemID] = &domain.Item{ID: itemID, Quantity: 0, Version: 1}
	orders.orders[id] = &domain.Order{ID: id, UserID: uuid.New(), ItemID: itemID, Status: domain.OrderStatusCreated, Quantity: 4}

	performedBy := uuid.New()
	if err := svc.Cancel(context.Background(), id, performedBy); err != nil {
		t.Fatalf("Cancel failed: %v", err)
	}
	if orders.orders[id].Status != domain.OrderStatusCancelled {
		t.Errorf("expected status=cancelled, got %v", orders.orders[id].Status)
	}
	if len(inventory.adjustments) != 1 || inventory.adjustments[0].QuantityChange != 4 {
		t.Error("expected inventory restoration of +4")
	}
}

func TestOrderCancel_AutoClosedOrder_ReturnsInvalidTransition(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	id := uuid.New()
	orders.orders[id] = &domain.Order{ID: id, UserID: uuid.New(), Status: domain.OrderStatusAutoClosed}

	if err := svc.Cancel(context.Background(), id, uuid.New()); err == nil {
		t.Fatal("expected invalid transition error")
	}
}

func TestOrderRefund_PaidToRefunded_RestoresInventory(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	id := uuid.New()
	itemID := uuid.New()
	items.items[itemID] = &domain.Item{ID: itemID, Quantity: 0, Version: 1}
	orders.orders[id] = &domain.Order{ID: id, UserID: uuid.New(), ItemID: itemID, Status: domain.OrderStatusPaid, Quantity: 2}

	if err := svc.Refund(context.Background(), id, uuid.New()); err != nil {
		t.Fatalf("Refund failed: %v", err)
	}
	if orders.orders[id].Status != domain.OrderStatusRefunded {
		t.Errorf("expected status=refunded, got %v", orders.orders[id].Status)
	}
	if len(inventory.adjustments) != 1 || inventory.adjustments[0].QuantityChange != 2 {
		t.Error("expected inventory restoration of +2")
	}
}

func TestOrderRefund_CreatedOrder_ReturnsInvalidTransition(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	id := uuid.New()
	orders.orders[id] = &domain.Order{ID: id, UserID: uuid.New(), Status: domain.OrderStatusCreated}

	if err := svc.Refund(context.Background(), id, uuid.New()); err == nil {
		t.Fatal("expected invalid transition error")
	}
}

func TestOrderAutoClose_ExpiredCreated_ClosedAndInventoryRestored(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	id := uuid.New()
	itemID := uuid.New()
	items.items[itemID] = &domain.Item{ID: itemID, Quantity: 0, Version: 1}
	orders.orders[id] = &domain.Order{
		ID:          id,
		UserID:      uuid.New(),
		ItemID:      itemID,
		Status:      domain.OrderStatusCreated,
		Quantity:    3,
		AutoCloseAt: time.Now().Add(-1 * time.Minute), // expired
	}

	count, err := svc.AutoCloseExpired(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("AutoCloseExpired failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 order auto-closed, got %d", count)
	}
	if orders.orders[id].Status != domain.OrderStatusAutoClosed {
		t.Errorf("expected status=auto_closed, got %v", orders.orders[id].Status)
	}
	if len(inventory.adjustments) != 1 || inventory.adjustments[0].QuantityChange != 3 {
		t.Error("expected inventory restoration of +3")
	}
	if len(inventory.snapshots) != 1 || inventory.snapshots[0].Quantity != 3 {
		t.Error("expected snapshot showing restored quantity 3")
	}
}

func TestOrderSplit_QuantitiesDontSum_ReturnsValidation(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	id := uuid.New()
	orders.orders[id] = &domain.Order{ID: id, UserID: uuid.New(), Status: domain.OrderStatusCreated, Quantity: 10}

	_, err := svc.Split(context.Background(), id, []int{3, 4}, nil) // sums to 7 != 10
	if err == nil {
		t.Fatal("expected validation error for mismatched quantities")
	}
}

func TestOrderSplit_TerminalOrder_ReturnsInvalidTransition(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	id := uuid.New()
	orders.orders[id] = &domain.Order{ID: id, UserID: uuid.New(), Status: domain.OrderStatusCancelled, Quantity: 10}

	_, err := svc.Split(context.Background(), id, []int{5, 5}, nil)
	if err == nil {
		t.Fatal("expected invalid transition error for terminal order")
	}
}

func TestOrderSplit_Success_OriginalCancelled_NoInventoryAdj(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	id := uuid.New()
	itemID := uuid.New()
	orders.orders[id] = &domain.Order{
		ID:       id,
		UserID:   uuid.New(),
		ItemID:   itemID,
		Status:   domain.OrderStatusCreated,
		Quantity: 10,
	}

	children, err := svc.Split(context.Background(), id, []int{6, 4}, nil)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
	if orders.orders[id].Status != domain.OrderStatusCancelled {
		t.Errorf("expected original status=cancelled, got %v", orders.orders[id].Status)
	}
	// No inventory adjustments for split (net = 0).
	if len(inventory.adjustments) != 0 {
		t.Errorf("expected no inventory adjustments for split, got %d", len(inventory.adjustments))
	}
}

func TestOrderMerge_DifferentItems_ReturnsValidation(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	userID := uuid.New()
	id1 := uuid.New()
	id2 := uuid.New()
	orders.orders[id1] = &domain.Order{ID: id1, UserID: userID, ItemID: uuid.New(), Status: domain.OrderStatusCreated, Quantity: 3}
	orders.orders[id2] = &domain.Order{ID: id2, UserID: userID, ItemID: uuid.New(), Status: domain.OrderStatusCreated, Quantity: 2}

	_, err := svc.Merge(context.Background(), []uuid.UUID{id1, id2}, nil)
	if err == nil {
		t.Fatal("expected validation error for different items")
	}
}

func TestOrderMerge_Success_OriginalsCancelled_NoInventoryAdj(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	userID := uuid.New()
	itemID := uuid.New()
	id1 := uuid.New()
	id2 := uuid.New()
	orders.orders[id1] = &domain.Order{ID: id1, UserID: userID, ItemID: itemID, Status: domain.OrderStatusCreated, Quantity: 3, UnitPrice: 50}
	orders.orders[id2] = &domain.Order{ID: id2, UserID: userID, ItemID: itemID, Status: domain.OrderStatusCreated, Quantity: 2, UnitPrice: 50}

	merged, err := svc.Merge(context.Background(), []uuid.UUID{id1, id2}, nil)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}
	if merged.Quantity != 5 {
		t.Errorf("expected merged quantity=5, got %d", merged.Quantity)
	}
	if orders.orders[id1].Status != domain.OrderStatusCancelled {
		t.Error("expected original 1 to be cancelled")
	}
	if orders.orders[id2].Status != domain.OrderStatusCancelled {
		t.Error("expected original 2 to be cancelled")
	}
	// No inventory adjustments for merge (net = 0).
	if len(inventory.adjustments) != 0 {
		t.Errorf("expected no inventory adjustments for merge, got %d", len(inventory.adjustments))
	}
}

func TestGetTimeline_ReturnsEntries(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	id := uuid.New()
	orders.orders[id] = &domain.Order{ID: id, UserID: uuid.New(), Status: domain.OrderStatusCreated}
	timelines.entries = []domain.OrderTimelineEntry{
		{ID: uuid.New(), OrderID: id, Action: "created"},
		{ID: uuid.New(), OrderID: id, Action: "paid"},
	}

	entries, err := svc.GetTimeline(context.Background(), id)
	if err != nil {
		t.Fatalf("GetTimeline failed: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 timeline entries, got %d", len(entries))
	}
}

func TestOrderGetForActor_RejectsNonOwnerWithoutManagePermission(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	orderID := uuid.New()
	ownerID := uuid.New()
	orders.orders[orderID] = &domain.Order{ID: orderID, UserID: ownerID, Status: domain.OrderStatusCreated}

	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleMember}
	if _, err := svc.GetForActor(context.Background(), actor, orderID); err != domain.ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestOrderCancelForActor_AllowsOwnerOnCreatedOrder(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	orderID := uuid.New()
	itemID := uuid.New()
	ownerID := uuid.New()
	items.items[itemID] = &domain.Item{ID: itemID, Quantity: 0, Version: 1}
	orders.orders[orderID] = &domain.Order{
		ID:       orderID,
		UserID:   ownerID,
		ItemID:   itemID,
		Status:   domain.OrderStatusCreated,
		Quantity: 2,
	}

	actor := &domain.User{ID: ownerID, Role: domain.UserRoleMember}
	if err := svc.CancelForActor(context.Background(), actor, orderID); err != nil {
		t.Fatalf("CancelForActor failed: %v", err)
	}
	if orders.orders[orderID].Status != domain.OrderStatusCancelled {
		t.Fatalf("expected order to be cancelled, got %s", orders.orders[orderID].Status)
	}
}

func TestSplitForActor_ForbiddenForNonManager(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	id := uuid.New()
	orders.orders[id] = &domain.Order{ID: id, UserID: uuid.New(), Status: domain.OrderStatusCreated, Quantity: 10}

	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleMember}
	if _, err := svc.SplitForActor(context.Background(), actor, id, []int{5, 5}, nil); err != domain.ErrForbidden {
		t.Fatalf("expected ErrForbidden for member role, got %v", err)
	}
}

func TestSplitForActor_SucceedsForManager(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	id := uuid.New()
	orders.orders[id] = &domain.Order{ID: id, UserID: uuid.New(), Status: domain.OrderStatusCreated, Quantity: 10}

	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleAdministrator}
	children, err := svc.SplitForActor(context.Background(), actor, id, []int{6, 4}, nil)
	if err != nil {
		t.Fatalf("SplitForActor failed for administrator: %v", err)
	}
	if len(children) != 2 {
		t.Errorf("expected 2 children, got %d", len(children))
	}
	if len(timelines.entries) < 3 {
		t.Fatalf("expected split to write timeline entries, got %d", len(timelines.entries))
	}
	for _, entry := range timelines.entries {
		if entry.Action == "split" && entry.PerformedBy != actor.ID {
			t.Fatalf("expected split timeline actor %s, got %s", actor.ID, entry.PerformedBy)
		}
	}
}

func TestMergeForActor_ForbiddenForNonManager(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	userID := uuid.New()
	itemID := uuid.New()
	id1 := uuid.New()
	id2 := uuid.New()
	orders.orders[id1] = &domain.Order{ID: id1, UserID: userID, ItemID: itemID, Status: domain.OrderStatusCreated, Quantity: 3}
	orders.orders[id2] = &domain.Order{ID: id2, UserID: userID, ItemID: itemID, Status: domain.OrderStatusCreated, Quantity: 2}

	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleMember}
	if _, err := svc.MergeForActor(context.Background(), actor, []uuid.UUID{id1, id2}, nil); err != domain.ErrForbidden {
		t.Fatalf("expected ErrForbidden for member role, got %v", err)
	}
}

func TestMergeForActor_SucceedsForManager(t *testing.T) {
	orders := newMockOrderRepo()
	timelines := newMockTimelineRepo()
	items := newMockItemRepo()
	inventory := newMockInventoryRepo()
	svc := newOrderService(orders, timelines, items, inventory)

	userID := uuid.New()
	itemID := uuid.New()
	id1 := uuid.New()
	id2 := uuid.New()
	orders.orders[id1] = &domain.Order{ID: id1, UserID: userID, ItemID: itemID, Status: domain.OrderStatusCreated, Quantity: 3}
	orders.orders[id2] = &domain.Order{ID: id2, UserID: userID, ItemID: itemID, Status: domain.OrderStatusCreated, Quantity: 2}

	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleAdministrator}
	merged, err := svc.MergeForActor(context.Background(), actor, []uuid.UUID{id1, id2}, nil)
	if err != nil {
		t.Fatalf("MergeForActor failed for administrator: %v", err)
	}
	if merged.Quantity != 5 {
		t.Errorf("expected merged quantity=5, got %d", merged.Quantity)
	}
	for _, entry := range timelines.entries {
		if entry.Action == "merged" && entry.PerformedBy != actor.ID {
			t.Fatalf("expected merged timeline actor %s, got %s", actor.ID, entry.PerformedBy)
		}
	}
}
