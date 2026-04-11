package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GroupBuyCampaign represents a group-buy campaign for an item.
type GroupBuyCampaign struct {
	ID                  uuid.UUID
	ItemID              uuid.UUID
	MinQuantity         int
	CurrentCommittedQty int
	CutoffTime          time.Time
	Status              CampaignStatus
	CreatedBy           uuid.UUID
	CreatedAt           time.Time
	EvaluatedAt         *time.Time
}

// IsAtCutoff returns true if the given time is at or past the campaign's cutoff time.
func (c *GroupBuyCampaign) IsAtCutoff(now time.Time) bool {
	return !now.Before(c.CutoffTime)
}

// MeetsThreshold returns true if the current committed quantity meets or exceeds
// the minimum quantity required for the campaign to succeed.
func (c *GroupBuyCampaign) MeetsThreshold() bool {
	return c.CurrentCommittedQty >= c.MinQuantity
}

// Evaluate evaluates the campaign at the given time. If the campaign is not active
// or the cutoff has not been reached, an error is returned. Otherwise, the status
// is set to succeeded if the threshold is met, or failed otherwise.
func (c *GroupBuyCampaign) Evaluate(now time.Time) error {
	if c.Status != CampaignStatusActive {
		return fmt.Errorf("cannot evaluate campaign in %s status", c.Status)
	}
	if !c.IsAtCutoff(now) {
		return fmt.Errorf("campaign cutoff time has not been reached")
	}

	if c.MeetsThreshold() {
		c.Status = CampaignStatusSucceeded
	} else {
		c.Status = CampaignStatusFailed
	}
	c.EvaluatedAt = &now
	return nil
}

// GroupBuyParticipant represents a user's participation in a group-buy campaign.
type GroupBuyParticipant struct {
	ID         uuid.UUID
	CampaignID uuid.UUID
	UserID     uuid.UUID
	Quantity   int
	OrderID    uuid.UUID
	JoinedAt   time.Time
}
