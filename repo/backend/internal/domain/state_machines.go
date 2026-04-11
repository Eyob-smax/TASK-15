package domain

// --- Order State Machine ---

// validOrderTransitions defines the allowed state transitions for orders.
var validOrderTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusCreated: {OrderStatusPaid, OrderStatusCancelled, OrderStatusAutoClosed},
	OrderStatusPaid:    {OrderStatusCancelled, OrderStatusRefunded},
}

// ValidOrderTransition returns true if the transition from one order status
// to another is permitted by the order state machine.
func ValidOrderTransition(from, to OrderStatus) bool {
	allowed, ok := validOrderTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// TransitionOrder attempts to transition an order to a new status.
// Returns ErrInvalidTransition if the transition is not allowed.
func TransitionOrder(order *Order, newStatus OrderStatus) error {
	if !ValidOrderTransition(order.Status, newStatus) {
		return &ErrInvalidTransition{
			Entity: "order",
			From:   string(order.Status),
			To:     string(newStatus),
		}
	}
	order.Status = newStatus
	return nil
}

// --- Campaign State Machine ---

// validCampaignTransitions defines the allowed state transitions for campaigns.
var validCampaignTransitions = map[CampaignStatus][]CampaignStatus{
	CampaignStatusActive: {CampaignStatusSucceeded, CampaignStatusFailed, CampaignStatusCancelled},
}

// ValidCampaignTransition returns true if the transition from one campaign status
// to another is permitted by the campaign state machine.
func ValidCampaignTransition(from, to CampaignStatus) bool {
	allowed, ok := validCampaignTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// TransitionCampaign attempts to transition a campaign to a new status.
// Returns ErrInvalidTransition if the transition is not allowed.
func TransitionCampaign(campaign *GroupBuyCampaign, newStatus CampaignStatus) error {
	if !ValidCampaignTransition(campaign.Status, newStatus) {
		return &ErrInvalidTransition{
			Entity: "campaign",
			From:   string(campaign.Status),
			To:     string(newStatus),
		}
	}
	campaign.Status = newStatus
	return nil
}

// --- Purchase Order State Machine ---

// validPOTransitions defines the allowed state transitions for purchase orders.
var validPOTransitions = map[POStatus][]POStatus{
	POStatusCreated:  {POStatusApproved},
	POStatusApproved: {POStatusReceived, POStatusVoided},
	POStatusReceived: {POStatusReturned, POStatusVoided},
}

// ValidPOTransition returns true if the transition from one purchase order status
// to another is permitted by the PO state machine.
func ValidPOTransition(from, to POStatus) bool {
	allowed, ok := validPOTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// TransitionPO attempts to transition a purchase order to a new status.
// Returns ErrInvalidTransition if the transition is not allowed.
func TransitionPO(po *PurchaseOrder, newStatus POStatus) error {
	if !ValidPOTransition(po.Status, newStatus) {
		return &ErrInvalidTransition{
			Entity: "purchase_order",
			From:   string(po.Status),
			To:     string(newStatus),
		}
	}
	po.Status = newStatus
	return nil
}
