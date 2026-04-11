package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
)

// CampaignServiceImpl implements CampaignService.
type CampaignServiceImpl struct {
	campaigns    store.CampaignRepository
	participants store.ParticipantRepository
	timelines    store.TimelineRepository
	items        store.ItemRepository
	availability store.AvailabilityWindowRepository
	blackouts    store.BlackoutWindowRepository
	orders       store.OrderRepository
	inventory    store.InventoryRepository
	audit        AuditService
	txPool       *pgxpool.Pool
}

// NewCampaignService creates a CampaignServiceImpl backed by the given repositories.
func NewCampaignService(
	campaigns store.CampaignRepository,
	participants store.ParticipantRepository,
	timelines store.TimelineRepository,
	items store.ItemRepository,
	availability store.AvailabilityWindowRepository,
	blackouts store.BlackoutWindowRepository,
	orders store.OrderRepository,
	inventory store.InventoryRepository,
	audit AuditService,
	txPool *pgxpool.Pool,
) *CampaignServiceImpl {
	return &CampaignServiceImpl{
		campaigns:    campaigns,
		participants: participants,
		timelines:    timelines,
		items:        items,
		availability: availability,
		blackouts:    blackouts,
		orders:       orders,
		inventory:    inventory,
		audit:        audit,
		txPool:       txPool,
	}
}

func (s *CampaignServiceImpl) Create(ctx context.Context, campaign *domain.GroupBuyCampaign) (*domain.GroupBuyCampaign, error) {
	now := time.Now().UTC()
	if err := withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		item, err := s.items.GetByID(txCtx, campaign.ItemID)
		if err != nil {
			return err
		}
		if item.Status != domain.ItemStatusPublished {
			return &domain.ErrValidation{Field: "item_id", Message: "campaign requires a published item"}
		}
		campaign.ID = uuid.New()
		campaign.Status = domain.CampaignStatusActive
		campaign.CreatedAt = now
		campaign.CurrentCommittedQty = 0
		if err := s.campaigns.Create(txCtx, campaign); err != nil {
			return err
		}
		return s.audit.Log(txCtx, "campaign.created", "campaign", campaign.ID, campaign.CreatedBy, map[string]interface{}{
			"item_id":      campaign.ItemID.String(),
			"min_quantity": campaign.MinQuantity,
			"cutoff_time":  campaign.CutoffTime,
		})
	}); err != nil {
		return nil, err
	}
	return campaign, nil
}

func (s *CampaignServiceImpl) Get(ctx context.Context, id uuid.UUID) (*domain.GroupBuyCampaign, error) {
	return s.campaigns.GetByID(ctx, id)
}

func (s *CampaignServiceImpl) List(ctx context.Context, page, pageSize int) ([]domain.GroupBuyCampaign, int, error) {
	return s.campaigns.List(ctx, page, pageSize)
}

func (s *CampaignServiceImpl) Join(ctx context.Context, campaignID, userID uuid.UUID, quantity int) (*domain.GroupBuyParticipant, error) {
	now := time.Now().UTC()
	var participant *domain.GroupBuyParticipant
	if err := withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		campaign, err := s.campaigns.GetByID(txCtx, campaignID)
		if err != nil {
			return err
		}
		if campaign.Status != domain.CampaignStatusActive {
			return &domain.ErrInvalidTransition{
				Entity: "campaign",
				From:   string(campaign.Status),
				To:     "joined",
			}
		}
		if campaign.IsAtCutoff(now) {
			return &domain.ErrValidation{Field: "campaign_id", Message: "campaign has passed its cutoff time"}
		}

		item, err := s.inventoryCoordinator().ensureReservable(txCtx, campaign.ItemID, quantity, now)
		if err != nil {
			return err
		}
		order := &domain.Order{
			ID:          uuid.New(),
			UserID:      userID,
			ItemID:      campaign.ItemID,
			CampaignID:  &campaignID,
			Quantity:    quantity,
			UnitPrice:   item.UnitPrice,
			TotalAmount: float64(quantity) * item.UnitPrice,
			Status:      domain.OrderStatusCreated,
			AutoCloseAt: now.Add(domain.AutoCloseTimeout),
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := s.orders.Create(txCtx, order); err != nil {
			return err
		}
		if err := s.timelines.Create(txCtx, &domain.OrderTimelineEntry{
			ID:          uuid.New(),
			OrderID:     order.ID,
			Action:      "created",
			Description: "order placed via campaign join",
			PerformedBy: userID,
			CreatedAt:   now,
		}); err != nil {
			return err
		}

		participant = &domain.GroupBuyParticipant{
			ID:         uuid.New(),
			CampaignID: campaignID,
			UserID:     userID,
			Quantity:   quantity,
			OrderID:    order.ID,
			JoinedAt:   now,
		}
		if err := s.participants.Create(txCtx, participant); err != nil {
			return err
		}
		committed, err := s.participants.CountCommittedQuantity(txCtx, campaignID)
		if err != nil {
			return err
		}
		campaign.CurrentCommittedQty = committed
		if err := s.campaigns.Update(txCtx, campaign); err != nil {
			return err
		}
		if err := s.inventoryCoordinator().applyChange(txCtx, campaign.ItemID, -quantity, "group-buy reservation", userID, now); err != nil {
			return err
		}
		return s.audit.Log(txCtx, "campaign.joined", "campaign", campaignID, userID, map[string]interface{}{
			"quantity": quantity,
			"order_id": order.ID.String(),
		})
	}); err != nil {
		return nil, err
	}
	return participant, nil
}

func (s *CampaignServiceImpl) EvaluateAtCutoff(ctx context.Context, id uuid.UUID, now time.Time, performedBy uuid.UUID) error {
	return withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		campaign, err := s.campaigns.GetByID(txCtx, id)
		if err != nil {
			return err
		}

		committed, err := s.participants.CountCommittedQuantity(txCtx, id)
		if err != nil {
			return err
		}
		campaign.CurrentCommittedQty = committed

		if err := campaign.Evaluate(now); err != nil {
			return err
		}
		if err := s.campaigns.Update(txCtx, campaign); err != nil {
			return err
		}
		if campaign.Status == domain.CampaignStatusFailed {
			participants, err := s.participants.ListByCampaign(txCtx, id)
			if err != nil {
				return err
			}
			for _, p := range participants {
				order, err := s.orders.GetByID(txCtx, p.OrderID)
				if err != nil {
					return err
				}
				if order.Status == domain.OrderStatusCreated {
					if err := domain.TransitionOrder(order, domain.OrderStatusAutoClosed); err != nil {
						return err
					}
					order.CancelledAt = &now
					order.UpdatedAt = now
					if err := s.orders.Update(txCtx, order); err != nil {
						return err
					}
					if err := s.inventoryCoordinator().applyChange(txCtx, order.ItemID, p.Quantity, "campaign failed - reservation released", p.UserID, now); err != nil {
						return err
					}
				}
			}
		}
		return s.audit.Log(txCtx, "campaign.evaluated", "campaign", campaign.ID, performedBy, map[string]interface{}{
			"status":    string(campaign.Status),
			"committed": campaign.CurrentCommittedQty,
			"min":       campaign.MinQuantity,
		})
	})
}

func (s *CampaignServiceImpl) Cancel(ctx context.Context, id uuid.UUID, performedBy uuid.UUID) error {
	return withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		campaign, err := s.campaigns.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		if err := domain.TransitionCampaign(campaign, domain.CampaignStatusCancelled); err != nil {
			return err
		}
		if err := s.campaigns.Update(txCtx, campaign); err != nil {
			return err
		}

		participants, err := s.participants.ListByCampaign(txCtx, id)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		for _, p := range participants {
			order, err := s.orders.GetByID(txCtx, p.OrderID)
			if err != nil {
				return err
			}
			if order.Status == domain.OrderStatusCreated {
				if err := domain.TransitionOrder(order, domain.OrderStatusCancelled); err != nil {
					return err
				}
				order.CancelledAt = &now
				order.UpdatedAt = now
				if err := s.orders.Update(txCtx, order); err != nil {
					return err
				}
				if err := s.inventoryCoordinator().applyChange(txCtx, order.ItemID, p.Quantity, "campaign cancelled - reservation released", p.UserID, now); err != nil {
					return err
				}
			}
		}
		return s.audit.Log(txCtx, "campaign.cancelled", "campaign", campaign.ID, performedBy, nil)
	})
}

func (s *CampaignServiceImpl) ListPastCutoff(ctx context.Context, now time.Time) ([]domain.GroupBuyCampaign, error) {
	return s.campaigns.ListDueCampaigns(ctx, now)
}

func (s *CampaignServiceImpl) inventoryCoordinator() inventoryCoordinator {
	return inventoryCoordinator{
		items:        s.items,
		inventory:    s.inventory,
		availability: s.availability,
		blackouts:    s.blackouts,
	}
}
