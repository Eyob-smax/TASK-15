package domain_test

import (
	"testing"
	"time"

	"fitcommerce/internal/domain"

	"github.com/google/uuid"
)

// ─── Order State Machine: ValidOrderTransition ─────────────────────────────────

func TestValidOrderTransition_CreatedToPaid(t *testing.T) {
	if !domain.ValidOrderTransition(domain.OrderStatusCreated, domain.OrderStatusPaid) {
		t.Error("expected created -> paid to be valid")
	}
}

func TestValidOrderTransition_CreatedToCancelled(t *testing.T) {
	if !domain.ValidOrderTransition(domain.OrderStatusCreated, domain.OrderStatusCancelled) {
		t.Error("expected created -> cancelled to be valid")
	}
}

func TestValidOrderTransition_CreatedToAutoClosed(t *testing.T) {
	if !domain.ValidOrderTransition(domain.OrderStatusCreated, domain.OrderStatusAutoClosed) {
		t.Error("expected created -> auto_closed to be valid")
	}
}

func TestValidOrderTransition_PaidToCancelled(t *testing.T) {
	if !domain.ValidOrderTransition(domain.OrderStatusPaid, domain.OrderStatusCancelled) {
		t.Error("expected paid -> cancelled to be valid")
	}
}

func TestValidOrderTransition_PaidToRefunded(t *testing.T) {
	if !domain.ValidOrderTransition(domain.OrderStatusPaid, domain.OrderStatusRefunded) {
		t.Error("expected paid -> refunded to be valid")
	}
}

func TestValidOrderTransition_PaidToCreated(t *testing.T) {
	if domain.ValidOrderTransition(domain.OrderStatusPaid, domain.OrderStatusCreated) {
		t.Error("expected paid -> created to be INVALID (backward transition)")
	}
}

func TestValidOrderTransition_CancelledToPaid(t *testing.T) {
	if domain.ValidOrderTransition(domain.OrderStatusCancelled, domain.OrderStatusPaid) {
		t.Error("expected cancelled -> paid to be INVALID (terminal state)")
	}
}

func TestValidOrderTransition_RefundedToAny(t *testing.T) {
	allStatuses := domain.AllOrderStatuses()
	for _, target := range allStatuses {
		if domain.ValidOrderTransition(domain.OrderStatusRefunded, target) {
			t.Errorf("expected refunded -> %s to be INVALID (terminal state)", target)
		}
	}
}

func TestValidOrderTransition_AutoClosedToAny(t *testing.T) {
	allStatuses := domain.AllOrderStatuses()
	for _, target := range allStatuses {
		if domain.ValidOrderTransition(domain.OrderStatusAutoClosed, target) {
			t.Errorf("expected auto_closed -> %s to be INVALID (terminal state)", target)
		}
	}
}

// ─── TransitionOrder ───────────────────────────────────────────────────────────

func TestTransitionOrder_ValidTransition(t *testing.T) {
	order := &domain.Order{
		ID:     uuid.New(),
		Status: domain.OrderStatusCreated,
	}
	err := domain.TransitionOrder(order, domain.OrderStatusPaid)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if order.Status != domain.OrderStatusPaid {
		t.Errorf("expected status to be %q, got %q", domain.OrderStatusPaid, order.Status)
	}
}

func TestTransitionOrder_InvalidTransition(t *testing.T) {
	order := &domain.Order{
		ID:     uuid.New(),
		Status: domain.OrderStatusCancelled,
	}
	err := domain.TransitionOrder(order, domain.OrderStatusPaid)
	if err == nil {
		t.Fatal("expected error for invalid transition")
	}
	// Verify it is an ErrInvalidTransition
	if _, ok := err.(*domain.ErrInvalidTransition); !ok {
		t.Errorf("expected *ErrInvalidTransition, got %T", err)
	}
}

// ─── Campaign State Machine: ValidCampaignTransition ───────────────────────────

func TestValidCampaignTransition_ActiveToSucceeded(t *testing.T) {
	if !domain.ValidCampaignTransition(domain.CampaignStatusActive, domain.CampaignStatusSucceeded) {
		t.Error("expected active -> succeeded to be valid")
	}
}

func TestValidCampaignTransition_ActiveToFailed(t *testing.T) {
	if !domain.ValidCampaignTransition(domain.CampaignStatusActive, domain.CampaignStatusFailed) {
		t.Error("expected active -> failed to be valid")
	}
}

func TestValidCampaignTransition_ActiveToCancelled(t *testing.T) {
	if !domain.ValidCampaignTransition(domain.CampaignStatusActive, domain.CampaignStatusCancelled) {
		t.Error("expected active -> cancelled to be valid")
	}
}

func TestValidCampaignTransition_SucceededToAny(t *testing.T) {
	allStatuses := domain.AllCampaignStatuses()
	for _, target := range allStatuses {
		if domain.ValidCampaignTransition(domain.CampaignStatusSucceeded, target) {
			t.Errorf("expected succeeded -> %s to be INVALID (terminal state)", target)
		}
	}
}

func TestValidCampaignTransition_FailedToAny(t *testing.T) {
	allStatuses := domain.AllCampaignStatuses()
	for _, target := range allStatuses {
		if domain.ValidCampaignTransition(domain.CampaignStatusFailed, target) {
			t.Errorf("expected failed -> %s to be INVALID (terminal state)", target)
		}
	}
}

// ─── TransitionCampaign ────────────────────────────────────────────────────────

func TestTransitionCampaign_ValidTransition(t *testing.T) {
	campaign := &domain.GroupBuyCampaign{
		ID:     uuid.New(),
		Status: domain.CampaignStatusActive,
	}
	err := domain.TransitionCampaign(campaign, domain.CampaignStatusSucceeded)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if campaign.Status != domain.CampaignStatusSucceeded {
		t.Errorf("expected status %q, got %q", domain.CampaignStatusSucceeded, campaign.Status)
	}
}

func TestTransitionCampaign_InvalidTransition(t *testing.T) {
	campaign := &domain.GroupBuyCampaign{
		ID:     uuid.New(),
		Status: domain.CampaignStatusFailed,
	}
	err := domain.TransitionCampaign(campaign, domain.CampaignStatusActive)
	if err == nil {
		t.Fatal("expected error for invalid campaign transition")
	}
	if _, ok := err.(*domain.ErrInvalidTransition); !ok {
		t.Errorf("expected *ErrInvalidTransition, got %T", err)
	}
}

// ─── PO State Machine: ValidPOTransition ───────────────────────────────────────

func TestValidPOTransition_CreatedToApproved(t *testing.T) {
	if !domain.ValidPOTransition(domain.POStatusCreated, domain.POStatusApproved) {
		t.Error("expected created -> approved to be valid")
	}
}

func TestValidPOTransition_ApprovedToReceived(t *testing.T) {
	if !domain.ValidPOTransition(domain.POStatusApproved, domain.POStatusReceived) {
		t.Error("expected approved -> received to be valid")
	}
}

func TestValidPOTransition_ApprovedToVoided(t *testing.T) {
	if !domain.ValidPOTransition(domain.POStatusApproved, domain.POStatusVoided) {
		t.Error("expected approved -> voided to be valid")
	}
}

func TestValidPOTransition_ReceivedToReturned(t *testing.T) {
	if !domain.ValidPOTransition(domain.POStatusReceived, domain.POStatusReturned) {
		t.Error("expected received -> returned to be valid")
	}
}

func TestValidPOTransition_ReceivedToVoided(t *testing.T) {
	if !domain.ValidPOTransition(domain.POStatusReceived, domain.POStatusVoided) {
		t.Error("expected received -> voided to be valid")
	}
}

func TestValidPOTransition_CreatedToReceived(t *testing.T) {
	if domain.ValidPOTransition(domain.POStatusCreated, domain.POStatusReceived) {
		t.Error("expected created -> received to be INVALID (skip step)")
	}
}

func TestValidPOTransition_VoidedToAny(t *testing.T) {
	allStatuses := domain.AllPOStatuses()
	for _, target := range allStatuses {
		if domain.ValidPOTransition(domain.POStatusVoided, target) {
			t.Errorf("expected voided -> %s to be INVALID (terminal state)", target)
		}
	}
}

// ─── TransitionPO ──────────────────────────────────────────────────────────────

func TestTransitionPO_ValidTransition(t *testing.T) {
	po := &domain.PurchaseOrder{
		ID:     uuid.New(),
		Status: domain.POStatusCreated,
	}
	err := domain.TransitionPO(po, domain.POStatusApproved)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if po.Status != domain.POStatusApproved {
		t.Errorf("expected status %q, got %q", domain.POStatusApproved, po.Status)
	}
}

func TestTransitionPO_InvalidTransition(t *testing.T) {
	po := &domain.PurchaseOrder{
		ID:     uuid.New(),
		Status: domain.POStatusVoided,
	}
	err := domain.TransitionPO(po, domain.POStatusCreated)
	if err == nil {
		t.Fatal("expected error for invalid PO transition")
	}
	if _, ok := err.(*domain.ErrInvalidTransition); !ok {
		t.Errorf("expected *ErrInvalidTransition, got %T", err)
	}
}

// ─── Order.ShouldAutoClose ─────────────────────────────────────────────────────

func TestOrder_ShouldAutoClose_Expired(t *testing.T) {
	order := &domain.Order{
		ID:          uuid.New(),
		Status:      domain.OrderStatusCreated,
		AutoCloseAt: time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC),
	}
	now := time.Date(2026, 4, 9, 10, 30, 0, 0, time.UTC) // past auto_close_at
	if !order.ShouldAutoClose(now) {
		t.Error("expected ShouldAutoClose to be true for expired order")
	}
}

func TestOrder_ShouldAutoClose_NotExpired(t *testing.T) {
	order := &domain.Order{
		ID:          uuid.New(),
		Status:      domain.OrderStatusCreated,
		AutoCloseAt: time.Date(2026, 4, 9, 11, 0, 0, 0, time.UTC),
	}
	now := time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC) // before auto_close_at
	if order.ShouldAutoClose(now) {
		t.Error("expected ShouldAutoClose to be false for non-expired order")
	}
}

func TestOrder_ShouldAutoClose_AlreadyPaid(t *testing.T) {
	order := &domain.Order{
		ID:          uuid.New(),
		Status:      domain.OrderStatusPaid,
		AutoCloseAt: time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC),
	}
	now := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC) // past auto_close_at
	if order.ShouldAutoClose(now) {
		t.Error("expected ShouldAutoClose to be false for paid order even if past deadline")
	}
}

// ─── Campaign.MeetsThreshold ───────────────────────────────────────────────────

func TestCampaign_MeetsThreshold(t *testing.T) {
	campaign := &domain.GroupBuyCampaign{
		ID:                  uuid.New(),
		MinQuantity:         10,
		CurrentCommittedQty: 15,
		Status:              domain.CampaignStatusActive,
	}
	if !campaign.MeetsThreshold() {
		t.Error("expected MeetsThreshold to be true when committed >= min")
	}
}

func TestCampaign_DoesNotMeetThreshold(t *testing.T) {
	campaign := &domain.GroupBuyCampaign{
		ID:                  uuid.New(),
		MinQuantity:         10,
		CurrentCommittedQty: 5,
		Status:              domain.CampaignStatusActive,
	}
	if campaign.MeetsThreshold() {
		t.Error("expected MeetsThreshold to be false when committed < min")
	}
}

// ─── Campaign.Evaluate ─────────────────────────────────────────────────────────

func TestCampaign_Evaluate_Success(t *testing.T) {
	cutoff := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	campaign := &domain.GroupBuyCampaign{
		ID:                  uuid.New(),
		MinQuantity:         10,
		CurrentCommittedQty: 12,
		CutoffTime:          cutoff,
		Status:              domain.CampaignStatusActive,
	}
	now := cutoff.Add(1 * time.Hour) // past cutoff
	err := campaign.Evaluate(now)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if campaign.Status != domain.CampaignStatusSucceeded {
		t.Errorf("expected status %q, got %q", domain.CampaignStatusSucceeded, campaign.Status)
	}
	if campaign.EvaluatedAt == nil {
		t.Error("expected EvaluatedAt to be set")
	}
}

func TestCampaign_Evaluate_Failure(t *testing.T) {
	cutoff := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	campaign := &domain.GroupBuyCampaign{
		ID:                  uuid.New(),
		MinQuantity:         10,
		CurrentCommittedQty: 3,
		CutoffTime:          cutoff,
		Status:              domain.CampaignStatusActive,
	}
	now := cutoff.Add(1 * time.Hour)
	err := campaign.Evaluate(now)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if campaign.Status != domain.CampaignStatusFailed {
		t.Errorf("expected status %q, got %q", domain.CampaignStatusFailed, campaign.Status)
	}
}

func TestCampaign_Evaluate_NotAtCutoff(t *testing.T) {
	cutoff := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	campaign := &domain.GroupBuyCampaign{
		ID:                  uuid.New(),
		MinQuantity:         10,
		CurrentCommittedQty: 15,
		CutoffTime:          cutoff,
		Status:              domain.CampaignStatusActive,
	}
	now := cutoff.Add(-24 * time.Hour) // before cutoff
	err := campaign.Evaluate(now)
	if err == nil {
		t.Fatal("expected error when evaluating before cutoff")
	}
}
