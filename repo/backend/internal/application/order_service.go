package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/security"
	"fitcommerce/internal/store"
)

// OrderServiceImpl implements OrderService.
type OrderServiceImpl struct {
	orders          store.OrderRepository
	timelines       store.TimelineRepository
	items           store.ItemRepository
	inventory       store.InventoryRepository
	availability    store.AvailabilityWindowRepository
	blackouts       store.BlackoutWindowRepository
	fulfillmentRepo store.FulfillmentRepository
	audit           AuditService
	txPool          *pgxpool.Pool
}

// NewOrderService creates an OrderServiceImpl backed by the given repositories.
func NewOrderService(
	orders store.OrderRepository,
	timelines store.TimelineRepository,
	items store.ItemRepository,
	inventory store.InventoryRepository,
	availability store.AvailabilityWindowRepository,
	blackouts store.BlackoutWindowRepository,
	fulfillmentRepo store.FulfillmentRepository,
	audit AuditService,
	txPool *pgxpool.Pool,
) *OrderServiceImpl {
	return &OrderServiceImpl{
		orders:          orders,
		timelines:       timelines,
		items:           items,
		inventory:       inventory,
		availability:    availability,
		blackouts:       blackouts,
		fulfillmentRepo: fulfillmentRepo,
		audit:           audit,
		txPool:          txPool,
	}
}

func (s *OrderServiceImpl) Create(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	now := time.Now().UTC()
	if err := withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		item, err := s.inventoryCoordinator().ensureReservable(txCtx, order.ItemID, order.Quantity, now)
		if err != nil {
			return err
		}

		order.ID = uuid.New()
		order.Status = domain.OrderStatusCreated
		order.UnitPrice = item.UnitPrice
		order.TotalAmount = float64(order.Quantity) * order.UnitPrice
		order.AutoCloseAt = now.Add(domain.AutoCloseTimeout)
		order.CreatedAt = now
		order.UpdatedAt = now

		if err := s.orders.Create(txCtx, order); err != nil {
			return err
		}
		if err := s.inventoryCoordinator().applyChange(txCtx, order.ItemID, -order.Quantity, "order reservation", order.UserID, now); err != nil {
			return err
		}
		if err := s.timelines.Create(txCtx, &domain.OrderTimelineEntry{
			ID:          uuid.New(),
			OrderID:     order.ID,
			Action:      "created",
			Description: "order placed",
			PerformedBy: order.UserID,
			CreatedAt:   now,
		}); err != nil {
			return err
		}
		return s.audit.Log(txCtx, "order.created", "order", order.ID, order.UserID, map[string]interface{}{
			"item_id":  order.ItemID.String(),
			"quantity": order.Quantity,
		})
	}); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderServiceImpl) Get(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return s.orders.GetByID(ctx, id)
}

func (s *OrderServiceImpl) GetForActor(ctx context.Context, actor *domain.User, id uuid.UUID) (*domain.Order, error) {
	order, err := s.orders.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := authorizeOrderReadActor(actor, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderServiceImpl) List(ctx context.Context, userID *uuid.UUID, page, pageSize int) ([]domain.Order, int, error) {
	return s.orders.List(ctx, userID, page, pageSize)
}

func (s *OrderServiceImpl) ListForActor(ctx context.Context, actor *domain.User, page, pageSize int) ([]domain.Order, int, error) {
	if actor == nil {
		return nil, 0, domain.ErrForbidden
	}
	if security.HasPermission(actor.Role, security.ActionManageOrders) {
		return s.orders.List(ctx, nil, page, pageSize)
	}
	return s.orders.List(ctx, &actor.ID, page, pageSize)
}

func (s *OrderServiceImpl) GetTimeline(ctx context.Context, orderID uuid.UUID) ([]domain.OrderTimelineEntry, error) {
	if _, err := s.orders.GetByID(ctx, orderID); err != nil {
		return nil, err
	}
	return s.timelines.ListByOrderID(ctx, orderID)
}

func (s *OrderServiceImpl) GetTimelineForActor(ctx context.Context, actor *domain.User, orderID uuid.UUID) ([]domain.OrderTimelineEntry, error) {
	if _, err := s.GetForActor(ctx, actor, orderID); err != nil {
		return nil, err
	}
	return s.timelines.ListByOrderID(ctx, orderID)
}

func (s *OrderServiceImpl) Pay(ctx context.Context, id uuid.UUID, settlementMarker string, performedBy uuid.UUID) error {
	now := time.Now().UTC()
	return withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		order, err := s.orders.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		if err := domain.TransitionOrder(order, domain.OrderStatusPaid); err != nil {
			return err
		}
		order.SettlementMarker = settlementMarker
		order.PaidAt = &now
		order.UpdatedAt = now
		if err := s.orders.Update(txCtx, order); err != nil {
			return err
		}
		if err := s.timelines.Create(txCtx, &domain.OrderTimelineEntry{
			ID:          uuid.New(),
			OrderID:     order.ID,
			Action:      "paid",
			Description: "payment recorded via settlement marker",
			PerformedBy: performedBy,
			CreatedAt:   now,
		}); err != nil {
			return err
		}
		return s.audit.Log(txCtx, "order.paid", "order", order.ID, performedBy, map[string]interface{}{
			"settlement_marker": settlementMarker,
		})
	})
}

func (s *OrderServiceImpl) Cancel(ctx context.Context, id uuid.UUID, performedBy uuid.UUID) error {
	now := time.Now().UTC()
	return withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		order, err := s.orders.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		if err := domain.TransitionOrder(order, domain.OrderStatusCancelled); err != nil {
			return err
		}
		order.CancelledAt = &now
		order.UpdatedAt = now
		if err := s.orders.Update(txCtx, order); err != nil {
			return err
		}
		if err := s.inventoryCoordinator().applyChange(txCtx, order.ItemID, order.Quantity, "order cancelled", performedBy, now); err != nil {
			return err
		}
		if err := s.timelines.Create(txCtx, &domain.OrderTimelineEntry{
			ID:          uuid.New(),
			OrderID:     order.ID,
			Action:      "cancelled",
			Description: "order cancelled",
			PerformedBy: performedBy,
			CreatedAt:   now,
		}); err != nil {
			return err
		}
		return s.audit.Log(txCtx, "order.cancelled", "order", order.ID, performedBy, nil)
	})
}

func (s *OrderServiceImpl) CancelForActor(ctx context.Context, actor *domain.User, id uuid.UUID) error {
	if actor == nil {
		return domain.ErrForbidden
	}

	order, err := s.orders.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := authorizeOrderCancellation(actor, order); err != nil {
		return err
	}
	return s.Cancel(ctx, id, actor.ID)
}

func (s *OrderServiceImpl) Refund(ctx context.Context, id uuid.UUID, performedBy uuid.UUID) error {
	now := time.Now().UTC()
	return withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		order, err := s.orders.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		if err := domain.TransitionOrder(order, domain.OrderStatusRefunded); err != nil {
			return err
		}
		order.RefundedAt = &now
		order.UpdatedAt = now
		if err := s.orders.Update(txCtx, order); err != nil {
			return err
		}
		if err := s.inventoryCoordinator().applyChange(txCtx, order.ItemID, order.Quantity, "order refunded", performedBy, now); err != nil {
			return err
		}
		if err := s.timelines.Create(txCtx, &domain.OrderTimelineEntry{
			ID:          uuid.New(),
			OrderID:     order.ID,
			Action:      "refunded",
			Description: "order refunded",
			PerformedBy: performedBy,
			CreatedAt:   now,
		}); err != nil {
			return err
		}
		return s.audit.Log(txCtx, "order.refunded", "order", order.ID, performedBy, nil)
	})
}

func (s *OrderServiceImpl) AddNote(ctx context.Context, id uuid.UUID, note string, performedBy uuid.UUID) error {
	now := time.Now().UTC()
	return withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		order, err := s.orders.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		order.Notes = note
		order.UpdatedAt = now
		if err := s.orders.Update(txCtx, order); err != nil {
			return err
		}
		if err := s.timelines.Create(txCtx, &domain.OrderTimelineEntry{
			ID:          uuid.New(),
			OrderID:     order.ID,
			Action:      "note_added",
			Description: "note added (content redacted)",
			PerformedBy: performedBy,
			CreatedAt:   now,
		}); err != nil {
			return err
		}
		return s.audit.Log(txCtx, "order.note_added", "order", order.ID, performedBy, map[string]interface{}{
			"note":         "[REDACTED]",
			"note_length":  len(note),
			"note_present": note != "",
		})
	})
}

func (s *OrderServiceImpl) Split(ctx context.Context, orderID uuid.UUID, quantities []int, fulfillment *FulfillmentInput) ([]domain.Order, error) {
	return s.splitWithActor(ctx, orderID, quantities, fulfillment, uuid.Nil)
}

func (s *OrderServiceImpl) splitWithActor(ctx context.Context, orderID uuid.UUID, quantities []int, fulfillment *FulfillmentInput, performedBy uuid.UUID) ([]domain.Order, error) {
	if len(quantities) < 2 {
		return nil, &domain.ErrValidation{Field: "quantities", Message: "split requires at least 2 quantities"}
	}

	var children []domain.Order

	err := withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		order, err := s.orders.GetByID(txCtx, orderID)
		if err != nil {
			return err
		}

		total := 0
		for _, q := range quantities {
			total += q
		}
		if total != order.Quantity {
			return &domain.ErrValidation{Field: "quantities", Message: "split quantities must sum to original order quantity"}
		}
		if order.Status.IsTerminal() {
			return &domain.ErrInvalidTransition{Entity: "order", From: string(order.Status), To: "split"}
		}

		now := time.Now().UTC()
		actorID := performedBy
		if actorID == uuid.Nil {
			actorID = order.UserID
		}
		children = make([]domain.Order, 0, len(quantities))
		childIDs := make([]string, 0, len(quantities))
		for _, qty := range quantities {
			child := domain.Order{
				ID:          uuid.New(),
				UserID:      order.UserID,
				ItemID:      order.ItemID,
				CampaignID:  order.CampaignID,
				Quantity:    qty,
				UnitPrice:   order.UnitPrice,
				TotalAmount: float64(qty) * order.UnitPrice,
				Status:      order.Status,
				AutoCloseAt: order.AutoCloseAt,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			if err := s.orders.Create(txCtx, &child); err != nil {
				return err
			}
			if err := s.timelines.Create(txCtx, &domain.OrderTimelineEntry{
				ID:          uuid.New(),
				OrderID:     child.ID,
				Action:      "split",
				Description: "created from order " + orderID.String(),
				PerformedBy: actorID,
				CreatedAt:   now,
			}); err != nil {
				return err
			}
			children = append(children, child)
			childIDs = append(childIDs, child.ID.String())
		}

		order.Status = domain.OrderStatusCancelled
		order.CancelledAt = &now
		order.UpdatedAt = now
		if err := s.orders.Update(txCtx, order); err != nil {
			return err
		}
		if err := s.timelines.Create(txCtx, &domain.OrderTimelineEntry{
			ID:          uuid.New(),
			OrderID:     order.ID,
			Action:      "split",
			Description: "split into " + itoa(len(children)) + " child orders",
			PerformedBy: actorID,
			CreatedAt:   now,
		}); err != nil {
			return err
		}
		if fulfillment != nil {
			group := &domain.FulfillmentGroup{
				ID:             uuid.New(),
				SupplierID:     fulfillment.SupplierID,
				WarehouseBinID: fulfillment.WarehouseBinID,
				PickupPoint:    fulfillment.PickupPoint,
				Status:         "pending",
				CreatedAt:      now,
			}
			if err := s.fulfillmentRepo.CreateGroup(txCtx, group); err != nil {
				return err
			}
			for _, child := range children {
				if err := s.fulfillmentRepo.AddGroupOrder(txCtx, &domain.FulfillmentGroupOrder{
					ID:                 uuid.New(),
					FulfillmentGroupID: group.ID,
					OrderID:            child.ID,
					Quantity:           child.Quantity,
				}); err != nil {
					return err
				}
			}
		}

		return s.audit.Log(txCtx, "order.split", "order", order.ID, actorID, map[string]interface{}{
			"child_ids": childIDs,
		})
	})
	if err != nil {
		return nil, err
	}
	return children, nil
}

func (s *OrderServiceImpl) Merge(ctx context.Context, orderIDs []uuid.UUID, fulfillment *FulfillmentInput) (*domain.Order, error) {
	return s.mergeWithActor(ctx, orderIDs, fulfillment, uuid.Nil)
}

func (s *OrderServiceImpl) mergeWithActor(ctx context.Context, orderIDs []uuid.UUID, fulfillment *FulfillmentInput, performedBy uuid.UUID) (*domain.Order, error) {
	if len(orderIDs) < 2 {
		return nil, &domain.ErrValidation{Field: "order_ids", Message: "merge requires at least 2 orders"}
	}

	var merged *domain.Order
	err := withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		orders := make([]*domain.Order, 0, len(orderIDs))
		for _, id := range orderIDs {
			o, err := s.orders.GetByID(txCtx, id)
			if err != nil {
				return err
			}
			orders = append(orders, o)
		}
		for _, o := range orders[1:] {
			if o.ItemID != orders[0].ItemID {
				return &domain.ErrValidation{Field: "order_ids", Message: "all orders must be for the same item"}
			}
			if o.UserID != orders[0].UserID {
				return &domain.ErrValidation{Field: "order_ids", Message: "all orders must belong to the same user"}
			}
		}
		for _, o := range orders {
			if o.Status.IsTerminal() {
				return &domain.ErrInvalidTransition{Entity: "order", From: string(o.Status), To: "merged"}
			}
		}

		totalQty := 0
		for _, o := range orders {
			totalQty += o.Quantity
		}

		var campaignID *uuid.UUID
		if orders[0].CampaignID != nil {
			same := true
			for _, o := range orders[1:] {
				if o.CampaignID == nil || *o.CampaignID != *orders[0].CampaignID {
					same = false
					break
				}
			}
			if same {
				campaignID = orders[0].CampaignID
			}
		}

		now := time.Now().UTC()
		actorID := performedBy
		if actorID == uuid.Nil {
			actorID = orders[0].UserID
		}
		merged = &domain.Order{
			ID:          uuid.New(),
			UserID:      orders[0].UserID,
			ItemID:      orders[0].ItemID,
			CampaignID:  campaignID,
			Quantity:    totalQty,
			UnitPrice:   orders[0].UnitPrice,
			TotalAmount: float64(totalQty) * orders[0].UnitPrice,
			Status:      domain.OrderStatusCreated,
			AutoCloseAt: now.Add(domain.AutoCloseTimeout),
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := s.orders.Create(txCtx, merged); err != nil {
			return err
		}

		sourceIDs := make([]string, 0, len(orders))
		for _, o := range orders {
			o.Status = domain.OrderStatusCancelled
			o.CancelledAt = &now
			o.UpdatedAt = now
			if err := s.orders.Update(txCtx, o); err != nil {
				return err
			}
			if err := s.timelines.Create(txCtx, &domain.OrderTimelineEntry{
				ID:          uuid.New(),
				OrderID:     o.ID,
				Action:      "merged",
				Description: "merged into order " + merged.ID.String(),
				PerformedBy: actorID,
				CreatedAt:   now,
			}); err != nil {
				return err
			}
			sourceIDs = append(sourceIDs, o.ID.String())
		}

		if err := s.timelines.Create(txCtx, &domain.OrderTimelineEntry{
			ID:          uuid.New(),
			OrderID:     merged.ID,
			Action:      "merged",
			Description: "merged from " + itoa(len(orders)) + " orders",
			PerformedBy: actorID,
			CreatedAt:   now,
		}); err != nil {
			return err
		}
		if fulfillment != nil {
			group := &domain.FulfillmentGroup{
				ID:             uuid.New(),
				SupplierID:     fulfillment.SupplierID,
				WarehouseBinID: fulfillment.WarehouseBinID,
				PickupPoint:    fulfillment.PickupPoint,
				Status:         "pending",
				CreatedAt:      now,
			}
			if err := s.fulfillmentRepo.CreateGroup(txCtx, group); err != nil {
				return err
			}
			if err := s.fulfillmentRepo.AddGroupOrder(txCtx, &domain.FulfillmentGroupOrder{
				ID:                 uuid.New(),
				FulfillmentGroupID: group.ID,
				OrderID:            merged.ID,
				Quantity:           merged.Quantity,
			}); err != nil {
				return err
			}
		}

		return s.audit.Log(txCtx, "order.merged", "order", merged.ID, actorID, map[string]interface{}{
			"source_ids": sourceIDs,
		})
	})
	if err != nil {
		return nil, err
	}
	return merged, nil
}

func (s *OrderServiceImpl) SplitForActor(ctx context.Context, actor *domain.User, orderID uuid.UUID, quantities []int, fulfillment *FulfillmentInput) ([]domain.Order, error) {
	if actor == nil || !security.HasPermission(actor.Role, security.ActionManageOrders) {
		return nil, domain.ErrForbidden
	}
	return s.splitWithActor(ctx, orderID, quantities, fulfillment, actor.ID)
}

func (s *OrderServiceImpl) MergeForActor(ctx context.Context, actor *domain.User, orderIDs []uuid.UUID, fulfillment *FulfillmentInput) (*domain.Order, error) {
	if actor == nil || !security.HasPermission(actor.Role, security.ActionManageOrders) {
		return nil, domain.ErrForbidden
	}
	return s.mergeWithActor(ctx, orderIDs, fulfillment, actor.ID)
}

func (s *OrderServiceImpl) AutoCloseExpired(ctx context.Context, now time.Time) (int, error) {
	expired, err := s.orders.ListExpiredUnpaid(ctx, now)
	if err != nil {
		return 0, err
	}

	count := 0
	for i := range expired {
		order := &expired[i]
		if err := domain.TransitionOrder(order, domain.OrderStatusAutoClosed); err != nil {
			return count, err
		}
		if err := withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
			order.CancelledAt = &now
			order.UpdatedAt = now
			if err := s.orders.Update(txCtx, order); err != nil {
				return err
			}
			if err := s.inventoryCoordinator().applyChange(txCtx, order.ItemID, order.Quantity, "auto-closed unpaid", uuid.Nil, now); err != nil {
				return err
			}
			if err := s.timelines.Create(txCtx, &domain.OrderTimelineEntry{
				ID:          uuid.New(),
				OrderID:     order.ID,
				Action:      "auto_closed",
				Description: "order auto-closed: unpaid timeout exceeded",
				PerformedBy: uuid.Nil,
				CreatedAt:   now,
			}); err != nil {
				return err
			}
			return s.audit.Log(txCtx, "order.auto_closed", "order", order.ID, uuid.Nil, nil)
		}); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *OrderServiceImpl) inventoryCoordinator() inventoryCoordinator {
	return inventoryCoordinator{
		items:        s.items,
		inventory:    s.inventory,
		availability: s.availability,
		blackouts:    s.blackouts,
	}
}

func authorizeOrderReadActor(actor *domain.User, order *domain.Order) error {
	if actor == nil {
		return domain.ErrForbidden
	}
	if security.HasPermission(actor.Role, security.ActionManageOrders) {
		return nil
	}
	if order.UserID != actor.ID {
		return domain.ErrForbidden
	}
	return nil
}

func authorizeOrderCancellation(actor *domain.User, order *domain.Order) error {
	if err := authorizeOrderReadActor(actor, order); err != nil {
		return err
	}
	if security.HasPermission(actor.Role, security.ActionManageOrders) {
		return nil
	}
	if order.Status != domain.OrderStatusCreated {
		return domain.ErrForbidden
	}
	return nil
}

// itoa converts an int to a string without importing strconv at package level.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
