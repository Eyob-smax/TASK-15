package campaign_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
)

func TestCampaignService_Get_Found(t *testing.T) {
	campaigns := newMockCampaignRepo()
	participants := newMockParticipantRepo()
	timelines := &mockTimelineRepo{}
	items := newMockItemRepo()
	orders := newMockOrderRepo()
	inventory := newMockInventoryRepo()
	svc := newCampaignService(campaigns, participants, timelines, items, orders, inventory)

	id := uuid.New()
	campaigns.campaigns[id] = &domain.GroupBuyCampaign{ID: id}

	got, err := svc.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.ID != id {
		t.Errorf("expected %v got %v", id, got.ID)
	}
}

func TestCampaignService_Get_NotFound(t *testing.T) {
	campaigns := newMockCampaignRepo()
	participants := newMockParticipantRepo()
	timelines := &mockTimelineRepo{}
	items := newMockItemRepo()
	orders := newMockOrderRepo()
	inventory := newMockInventoryRepo()
	svc := newCampaignService(campaigns, participants, timelines, items, orders, inventory)

	_, err := svc.Get(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestCampaignService_List(t *testing.T) {
	campaigns := newMockCampaignRepo()
	participants := newMockParticipantRepo()
	timelines := &mockTimelineRepo{}
	items := newMockItemRepo()
	orders := newMockOrderRepo()
	inventory := newMockInventoryRepo()
	svc := newCampaignService(campaigns, participants, timelines, items, orders, inventory)

	for i := 0; i < 3; i++ {
		id := uuid.New()
		campaigns.campaigns[id] = &domain.GroupBuyCampaign{ID: id, Status: domain.CampaignStatusActive}
	}

	rows, total, err := svc.List(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 3 || len(rows) != 3 {
		t.Errorf("expected 3, got total=%d rows=%d", total, len(rows))
	}
}

func TestCampaignService_ListPastCutoff_ReturnsDue(t *testing.T) {
	campaigns := newMockCampaignRepo()
	participants := newMockParticipantRepo()
	timelines := &mockTimelineRepo{}
	items := newMockItemRepo()
	orders := newMockOrderRepo()
	inventory := newMockInventoryRepo()
	svc := newCampaignService(campaigns, participants, timelines, items, orders, inventory)

	past := time.Now().Add(-2 * time.Hour)
	future := time.Now().Add(2 * time.Hour)

	pastID := uuid.New()
	futureID := uuid.New()
	campaigns.campaigns[pastID] = &domain.GroupBuyCampaign{ID: pastID, Status: domain.CampaignStatusActive, CutoffTime: past}
	campaigns.campaigns[futureID] = &domain.GroupBuyCampaign{ID: futureID, Status: domain.CampaignStatusActive, CutoffTime: future}

	due, err := svc.ListPastCutoff(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("ListPastCutoff failed: %v", err)
	}
	if len(due) != 1 {
		t.Fatalf("expected 1 due campaign, got %d", len(due))
	}
	if due[0].ID != pastID {
		t.Errorf("expected pastID, got %v", due[0].ID)
	}
}

func TestCampaignService_Cancel_Success_CancelsParticipantOrders(t *testing.T) {
	campaigns := newMockCampaignRepo()
	participants := newMockParticipantRepo()
	timelines := &mockTimelineRepo{}
	items := newMockItemRepo()
	orders := newMockOrderRepo()
	inventory := newMockInventoryRepo()
	svc := newCampaignService(campaigns, participants, timelines, items, orders, inventory)

	itemID := uuid.New()
	items.items[itemID] = &domain.Item{ID: itemID, Quantity: 0, Version: 1}

	campaignID := uuid.New()
	campaigns.campaigns[campaignID] = &domain.GroupBuyCampaign{
		ID:     campaignID,
		ItemID: itemID,
		Status: domain.CampaignStatusActive,
	}

	orderID := uuid.New()
	userID := uuid.New()
	orders.orders[orderID] = &domain.Order{
		ID:       orderID,
		UserID:   userID,
		ItemID:   itemID,
		Quantity: 4,
		Status:   domain.OrderStatusCreated,
	}
	participants.byCampaign[campaignID] = []*domain.GroupBuyParticipant{
		{ID: uuid.New(), CampaignID: campaignID, UserID: userID, OrderID: orderID, Quantity: 4},
	}

	if err := svc.Cancel(context.Background(), campaignID, uuid.New()); err != nil {
		t.Fatalf("Cancel failed: %v", err)
	}
	if campaigns.campaigns[campaignID].Status != domain.CampaignStatusCancelled {
		t.Errorf("expected cancelled, got %v", campaigns.campaigns[campaignID].Status)
	}
	if orders.orders[orderID].Status != domain.OrderStatusCancelled {
		t.Errorf("expected order cancelled, got %v", orders.orders[orderID].Status)
	}
	if len(inventory.adjustments) != 1 || inventory.adjustments[0].QuantityChange != 4 {
		t.Error("expected inventory +4 restoration")
	}
}

func TestCampaignService_Cancel_NotFound(t *testing.T) {
	campaigns := newMockCampaignRepo()
	participants := newMockParticipantRepo()
	timelines := &mockTimelineRepo{}
	items := newMockItemRepo()
	orders := newMockOrderRepo()
	inventory := newMockInventoryRepo()
	svc := newCampaignService(campaigns, participants, timelines, items, orders, inventory)

	err := svc.Cancel(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected error for missing campaign")
	}
}
