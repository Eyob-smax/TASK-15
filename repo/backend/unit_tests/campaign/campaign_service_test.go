package campaign_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
)

// --- Mock repositories ---

type mockCampaignRepo struct {
	campaigns map[uuid.UUID]*domain.GroupBuyCampaign
}

func newMockCampaignRepo() *mockCampaignRepo {
	return &mockCampaignRepo{campaigns: make(map[uuid.UUID]*domain.GroupBuyCampaign)}
}

func (m *mockCampaignRepo) Create(_ context.Context, c *domain.GroupBuyCampaign) error {
	m.campaigns[c.ID] = c
	return nil
}
func (m *mockCampaignRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.GroupBuyCampaign, error) {
	c, ok := m.campaigns[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return c, nil
}
func (m *mockCampaignRepo) List(_ context.Context, _, _ int) ([]domain.GroupBuyCampaign, int, error) {
	list := make([]domain.GroupBuyCampaign, 0, len(m.campaigns))
	for _, v := range m.campaigns {
		list = append(list, *v)
	}
	return list, len(list), nil
}
func (m *mockCampaignRepo) Update(_ context.Context, c *domain.GroupBuyCampaign) error {
	m.campaigns[c.ID] = c
	return nil
}
func (m *mockCampaignRepo) ListActive(_ context.Context) ([]domain.GroupBuyCampaign, error) {
	return nil, nil
}
func (m *mockCampaignRepo) ListDueCampaigns(_ context.Context, now time.Time) ([]domain.GroupBuyCampaign, error) {
	var result []domain.GroupBuyCampaign
	for _, c := range m.campaigns {
		if c.Status == domain.CampaignStatusActive && !now.Before(c.CutoffTime) {
			result = append(result, *c)
		}
	}
	return result, nil
}

type mockParticipantRepo struct {
	participants map[uuid.UUID]*domain.GroupBuyParticipant // keyed by participant ID
	byCampaign  map[uuid.UUID][]*domain.GroupBuyParticipant
}

func newMockParticipantRepo() *mockParticipantRepo {
	return &mockParticipantRepo{
		participants: make(map[uuid.UUID]*domain.GroupBuyParticipant),
		byCampaign:  make(map[uuid.UUID][]*domain.GroupBuyParticipant),
	}
}

func (m *mockParticipantRepo) Create(_ context.Context, p *domain.GroupBuyParticipant) error {
	m.participants[p.ID] = p
	m.byCampaign[p.CampaignID] = append(m.byCampaign[p.CampaignID], p)
	return nil
}
func (m *mockParticipantRepo) ListByCampaign(_ context.Context, campaignID uuid.UUID) ([]domain.GroupBuyParticipant, error) {
	var result []domain.GroupBuyParticipant
	for _, p := range m.byCampaign[campaignID] {
		result = append(result, *p)
	}
	return result, nil
}
func (m *mockParticipantRepo) CountCommittedQuantity(_ context.Context, campaignID uuid.UUID) (int, error) {
	total := 0
	for _, p := range m.byCampaign[campaignID] {
		total += p.Quantity
	}
	return total, nil
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
	return nil, 0, nil
}
func (m *mockOrderRepo) Update(_ context.Context, o *domain.Order) error {
	m.orders[o.ID] = o
	return nil
}
func (m *mockOrderRepo) ListExpiredUnpaid(_ context.Context, _ time.Time) ([]domain.Order, error) {
	return nil, nil
}

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

type mockAvailabilityRepo struct{}

func (m *mockAvailabilityRepo) Create(_ context.Context, _ *domain.AvailabilityWindow) error { return nil }
func (m *mockAvailabilityRepo) ListByItemID(_ context.Context, _ uuid.UUID) ([]domain.AvailabilityWindow, error) {
	return nil, nil
}
func (m *mockAvailabilityRepo) DeleteByItemID(_ context.Context, _ uuid.UUID) error { return nil }

type mockBlackoutRepo struct{}

func (m *mockBlackoutRepo) Create(_ context.Context, _ *domain.BlackoutWindow) error { return nil }
func (m *mockBlackoutRepo) ListByItemID(_ context.Context, _ uuid.UUID) ([]domain.BlackoutWindow, error) {
	return nil, nil
}
func (m *mockBlackoutRepo) DeleteByItemID(_ context.Context, _ uuid.UUID) error { return nil }

type mockTimelineRepo struct {
	entries []domain.OrderTimelineEntry
}

func (m *mockTimelineRepo) Create(_ context.Context, entry *domain.OrderTimelineEntry) error {
	m.entries = append(m.entries, *entry)
	return nil
}

func (m *mockTimelineRepo) ListByOrderID(_ context.Context, orderID uuid.UUID) ([]domain.OrderTimelineEntry, error) {
	result := make([]domain.OrderTimelineEntry, 0)
	for _, entry := range m.entries {
		if entry.OrderID == orderID {
			result = append(result, entry)
		}
	}
	return result, nil
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

func newCampaignService(campaigns *mockCampaignRepo, participants *mockParticipantRepo, timelines *mockTimelineRepo, items *mockItemRepo, orders *mockOrderRepo, inventory *mockInventoryRepo) *application.CampaignServiceImpl {
	return application.NewCampaignService(campaigns, participants, timelines, items, &mockAvailabilityRepo{}, &mockBlackoutRepo{}, orders, inventory, &mockAuditService{}, nil)
}

// --- Tests ---

func TestCampaignCreate_NonPublishedItem_ReturnsValidation(t *testing.T) {
	campaigns := newMockCampaignRepo()
	participants := newMockParticipantRepo()
	timelines := &mockTimelineRepo{}
	items := newMockItemRepo()
	orders := newMockOrderRepo()
	inventory := newMockInventoryRepo()

	itemID := uuid.New()
	items.items[itemID] = &domain.Item{
		ID:     itemID,
		Status: domain.ItemStatusDraft, // not published
	}
	svc := newCampaignService(campaigns, participants, timelines, items, orders, inventory)

	c := &domain.GroupBuyCampaign{
		ItemID:      itemID,
		MinQuantity: 10,
		CutoffTime:  time.Now().Add(24 * time.Hour),
		CreatedBy:   uuid.New(),
	}
	_, err := svc.Create(context.Background(), c)
	if err == nil {
		t.Fatal("expected validation error for non-published item")
	}
}

func TestCampaignCreate_PublishedItem_StatusActive(t *testing.T) {
	campaigns := newMockCampaignRepo()
	participants := newMockParticipantRepo()
	timelines := &mockTimelineRepo{}
	items := newMockItemRepo()
	orders := newMockOrderRepo()
	inventory := newMockInventoryRepo()

	itemID := uuid.New()
	items.items[itemID] = &domain.Item{
		ID:     itemID,
		Status: domain.ItemStatusPublished,
	}
	svc := newCampaignService(campaigns, participants, timelines, items, orders, inventory)

	c := &domain.GroupBuyCampaign{
		ItemID:      itemID,
		MinQuantity: 10,
		CutoffTime:  time.Now().Add(24 * time.Hour),
		CreatedBy:   uuid.New(),
	}
	created, err := svc.Create(context.Background(), c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.Status != domain.CampaignStatusActive {
		t.Errorf("expected status=active, got %v", created.Status)
	}
}

func TestCampaignJoin_CancelledCampaign_ReturnsInvalidTransition(t *testing.T) {
	campaigns := newMockCampaignRepo()
	participants := newMockParticipantRepo()
	timelines := &mockTimelineRepo{}
	items := newMockItemRepo()
	orders := newMockOrderRepo()
	inventory := newMockInventoryRepo()
	svc := newCampaignService(campaigns, participants, timelines, items, orders, inventory)

	campaignID := uuid.New()
	campaigns.campaigns[campaignID] = &domain.GroupBuyCampaign{
		ID:     campaignID,
		Status: domain.CampaignStatusCancelled,
	}

	_, err := svc.Join(context.Background(), campaignID, uuid.New(), 5)
	if err == nil {
		t.Fatal("expected invalid transition error")
	}
}

func TestCampaignJoin_PastCutoff_ReturnsValidation(t *testing.T) {
	campaigns := newMockCampaignRepo()
	participants := newMockParticipantRepo()
	timelines := &mockTimelineRepo{}
	items := newMockItemRepo()
	orders := newMockOrderRepo()
	inventory := newMockInventoryRepo()
	svc := newCampaignService(campaigns, participants, timelines, items, orders, inventory)

	itemID := uuid.New()
	items.items[itemID] = &domain.Item{
		ID:                itemID,
		Status:            domain.ItemStatusPublished,
		UnitPrice:         79,
		RefundableDeposit: 50,
		Quantity:          10,
		Version:           1,
	}
	campaignID := uuid.New()
	campaigns.campaigns[campaignID] = &domain.GroupBuyCampaign{
		ID:         campaignID,
		ItemID:     itemID,
		Status:     domain.CampaignStatusActive,
		CutoffTime: time.Now().Add(-1 * time.Hour), // past cutoff
	}

	_, err := svc.Join(context.Background(), campaignID, uuid.New(), 5)
	if err == nil {
		t.Fatal("expected validation error for past-cutoff campaign")
	}
}

func TestCampaignJoin_Active_CreatesParticipantAndInventoryAdj(t *testing.T) {
	campaigns := newMockCampaignRepo()
	participants := newMockParticipantRepo()
	timelines := &mockTimelineRepo{}
	items := newMockItemRepo()
	orders := newMockOrderRepo()
	inventory := newMockInventoryRepo()
	svc := newCampaignService(campaigns, participants, timelines, items, orders, inventory)

	itemID := uuid.New()
	items.items[itemID] = &domain.Item{
		ID:                itemID,
		Status:            domain.ItemStatusPublished,
		UnitPrice:         120,
		RefundableDeposit: 50,
		Quantity:          10,
		Version:           1,
	}
	campaignID := uuid.New()
	campaigns.campaigns[campaignID] = &domain.GroupBuyCampaign{
		ID:         campaignID,
		ItemID:     itemID,
		Status:     domain.CampaignStatusActive,
		CutoffTime: time.Now().Add(24 * time.Hour),
	}

	userID := uuid.New()
	_, err := svc.Join(context.Background(), campaignID, userID, 5)
	if err != nil {
		t.Fatalf("Join failed: %v", err)
	}

	// Participant created.
	if len(participants.byCampaign[campaignID]) != 1 {
		t.Error("expected 1 participant")
	}
	// Order created.
	if len(orders.orders) != 1 {
		t.Error("expected 1 order created")
	}
	for _, createdOrder := range orders.orders {
		if createdOrder.UnitPrice != 120 {
			t.Fatalf("expected campaign order unit price to use item.unit_price=120, got %v", createdOrder.UnitPrice)
		}
	}
	if len(timelines.entries) != 1 || timelines.entries[0].Action != "created" {
		t.Fatalf("expected campaign join to create one order timeline entry, got %#v", timelines.entries)
	}
	// Inventory decremented.
	if len(inventory.adjustments) != 1 {
		t.Error("expected 1 inventory adjustment")
	}
	if inventory.adjustments[0].QuantityChange != -5 {
		t.Errorf("expected adjustment=-5, got %d", inventory.adjustments[0].QuantityChange)
	}
	if len(inventory.snapshots) != 1 || inventory.snapshots[0].Quantity != 5 {
		t.Error("expected snapshot showing remaining quantity 5")
	}
}

func TestCampaignEvaluate_MetThreshold_Succeeded(t *testing.T) {
	campaigns := newMockCampaignRepo()
	participants := newMockParticipantRepo()
	timelines := &mockTimelineRepo{}
	items := newMockItemRepo()
	orders := newMockOrderRepo()
	inventory := newMockInventoryRepo()
	svc := newCampaignService(campaigns, participants, timelines, items, orders, inventory)

	itemID := uuid.New()
	campaignID := uuid.New()
	cutoff := time.Now().Add(-1 * time.Minute) // past cutoff
	campaigns.campaigns[campaignID] = &domain.GroupBuyCampaign{
		ID:                  campaignID,
		ItemID:              itemID,
		Status:              domain.CampaignStatusActive,
		CutoffTime:          cutoff,
		MinQuantity:         10,
		CurrentCommittedQty: 0,
		CreatedBy:           uuid.New(),
	}

	orderID := uuid.New()
	userID := uuid.New()
	orders.orders[orderID] = &domain.Order{
		ID:       orderID,
		UserID:   userID,
		ItemID:   itemID,
		Status:   domain.OrderStatusCreated,
		Quantity: 15,
	}
	participants.byCampaign[campaignID] = []*domain.GroupBuyParticipant{
		{ID: uuid.New(), CampaignID: campaignID, UserID: userID, OrderID: orderID, Quantity: 15},
	}

	now := time.Now()
	if err := svc.EvaluateAtCutoff(context.Background(), campaignID, now); err != nil {
		t.Fatalf("EvaluateAtCutoff failed: %v", err)
	}
	if campaigns.campaigns[campaignID].Status != domain.CampaignStatusSucceeded {
		t.Errorf("expected status=succeeded, got %v", campaigns.campaigns[campaignID].Status)
	}
}

func TestCampaignEvaluate_FailedThreshold_AutoClosesOrders(t *testing.T) {
	campaigns := newMockCampaignRepo()
	participants := newMockParticipantRepo()
	timelines := &mockTimelineRepo{}
	items := newMockItemRepo()
	orders := newMockOrderRepo()
	inventory := newMockInventoryRepo()
	svc := newCampaignService(campaigns, participants, timelines, items, orders, inventory)

	itemID := uuid.New()
	campaignID := uuid.New()
	cutoff := time.Now().Add(-1 * time.Minute)
	campaigns.campaigns[campaignID] = &domain.GroupBuyCampaign{
		ID:                  campaignID,
		ItemID:              itemID,
		Status:              domain.CampaignStatusActive,
		CutoffTime:          cutoff,
		MinQuantity:         100, // high threshold — will fail
		CurrentCommittedQty: 5,
		CreatedBy:           uuid.New(),
	}

	// Add a participating order.
	orderID := uuid.New()
	userID := uuid.New()
	items.items[itemID] = &domain.Item{ID: itemID, Quantity: 0, Version: 1}
	orders.orders[orderID] = &domain.Order{
		ID:     orderID,
		UserID: userID,
		ItemID: itemID,
		Status: domain.OrderStatusCreated,
		Quantity: 5,
	}
	participants.byCampaign[campaignID] = []*domain.GroupBuyParticipant{
		{ID: uuid.New(), CampaignID: campaignID, UserID: userID, OrderID: orderID, Quantity: 5},
	}

	now := time.Now()
	if err := svc.EvaluateAtCutoff(context.Background(), campaignID, now); err != nil {
		t.Fatalf("EvaluateAtCutoff failed: %v", err)
	}
	if campaigns.campaigns[campaignID].Status != domain.CampaignStatusFailed {
		t.Errorf("expected status=failed, got %v", campaigns.campaigns[campaignID].Status)
	}
	// Order should be auto-closed.
	if orders.orders[orderID].Status != domain.OrderStatusAutoClosed {
		t.Errorf("expected order status=auto_closed, got %v", orders.orders[orderID].Status)
	}
	// Inventory should be restored.
	if len(inventory.adjustments) != 1 || inventory.adjustments[0].QuantityChange != 5 {
		t.Error("expected inventory restoration adjustment of +5")
	}
}
