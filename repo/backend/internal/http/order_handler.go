package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
	"fitcommerce/internal/http/dto"
	"fitcommerce/internal/security"
)

// OrderHandler handles HTTP requests for the order endpoints.
type OrderHandler struct {
	svc application.OrderService
}

// NewOrderHandler creates an OrderHandler backed by the given service.
func NewOrderHandler(svc application.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// CreateOrder handles POST /api/v1/orders.
func (h *OrderHandler) CreateOrder(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	var req dto.CreateOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", err.Error()))
	}

	itemID, err := uuid.Parse(req.ItemID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid item_id"))
	}

	order := &domain.Order{
		UserID:   user.ID,
		ItemID:   itemID,
		Quantity: req.Quantity,
	}
	if req.CampaignID != nil {
		cid, err := uuid.Parse(*req.CampaignID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid campaign_id"))
		}
		order.CampaignID = &cid
	}

	created, err := h.svc.Create(c.Request().Context(), order)
	if err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusCreated, dto.SuccessResponse{Data: toOrderResponse(created)})
}

// ListOrders handles GET /api/v1/orders.
// ManageOrders roles see all orders; others see only their own.
func (h *OrderHandler) ListOrders(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	page, pageSize := parsePagination(c)

	orders, total, err := h.svc.ListForActor(c.Request().Context(), user, page, pageSize)
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.OrderResponse, len(orders))
	for i := range orders {
		resp[i] = toOrderResponse(&orders[i])
	}
	return c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       resp,
		Pagination: paginationMeta(page, pageSize, total),
	})
}

// GetOrder handles GET /api/v1/orders/:id.
// Non-ManageOrders users may only view their own orders.
func (h *OrderHandler) GetOrder(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid order id"))
	}

	order, err := h.svc.GetForActor(c.Request().Context(), user, id)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: toOrderResponse(order)})
}

// PayOrder handles POST /api/v1/orders/:id/pay.
func (h *OrderHandler) PayOrder(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid order id"))
	}

	var req dto.PayOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}

	if err := h.svc.Pay(c.Request().Context(), id, req.SettlementMarker, user.ID); err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: map[string]string{"message": "order paid"}})
}

// CancelOrder handles POST /api/v1/orders/:id/cancel.
// Cancellation policy:
//   - Members may cancel their own order only if it is in Created (unpaid) status.
//   - ManageOrders roles may cancel any order in Created or Paid status.
func (h *OrderHandler) CancelOrder(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid order id"))
	}

	if err := h.svc.CancelForActor(c.Request().Context(), user, id); err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: map[string]string{"message": "order cancelled"}})
}

// RefundOrder handles POST /api/v1/orders/:id/refund.
func (h *OrderHandler) RefundOrder(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid order id"))
	}

	if err := h.svc.Refund(c.Request().Context(), id, user.ID); err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: map[string]string{"message": "order refunded"}})
}

// AddOrderNote handles POST /api/v1/orders/:id/notes.
func (h *OrderHandler) AddOrderNote(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid order id"))
	}

	var req dto.AddOrderNoteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}

	if err := h.svc.AddNote(c.Request().Context(), id, req.Note, user.ID); err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: map[string]string{"message": "note added"}})
}

// GetOrderTimeline handles GET /api/v1/orders/:id/timeline.
func (h *OrderHandler) GetOrderTimeline(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid order id"))
	}

	entries, err := h.svc.GetTimelineForActor(c.Request().Context(), user, id)
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.TimelineEntryResponse, len(entries))
	for i, e := range entries {
		resp[i] = dto.TimelineEntryResponse{
			ID:          e.ID.String(),
			OrderID:     e.OrderID.String(),
			Action:      e.Action,
			Description: e.Description,
			PerformedBy: e.PerformedBy.String(),
			CreatedAt:   e.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: resp})
}

// SplitOrder handles POST /api/v1/orders/:id/split.
func (h *OrderHandler) SplitOrder(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid order id"))
	}

	var req dto.SplitOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}

	fulfillment, errMsg, parseErr := parseFulfillmentInput(req.SupplierID, req.WarehouseBinID, req.PickupPoint)
	if parseErr != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", errMsg))
	}

	children, err := h.svc.SplitForActor(c.Request().Context(), user, id, req.Quantities, fulfillment)
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.OrderResponse, len(children))
	for i := range children {
		resp[i] = toOrderResponse(&children[i])
	}
	return c.JSON(http.StatusCreated, dto.SuccessResponse{Data: resp})
}

// MergeOrders handles POST /api/v1/orders/merge.
func (h *OrderHandler) MergeOrders(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	var req dto.MergeOrdersRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}

	orderIDs := make([]uuid.UUID, len(req.OrderIDs))
	for i, s := range req.OrderIDs {
		oid, err := uuid.Parse(s)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid order_id: "+s))
		}
		orderIDs[i] = oid
	}

	fulfillment, errMsg, parseErr := parseFulfillmentInput(req.SupplierID, req.WarehouseBinID, req.PickupPoint)
	if parseErr != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", errMsg))
	}

	merged, err := h.svc.MergeForActor(c.Request().Context(), user, orderIDs, fulfillment)
	if err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusCreated, dto.SuccessResponse{Data: toOrderResponse(merged)})
}

// parseFulfillmentInput builds a FulfillmentInput from optional request fields.
// Returns (nil, "", nil) when no fulfillment metadata was supplied.
// Returns (nil, errorMessage, err) when a UUID field is malformed.
func parseFulfillmentInput(supplierID, warehouseBinID *string, pickupPoint string) (*application.FulfillmentInput, string, error) {
	if supplierID == nil && warehouseBinID == nil && pickupPoint == "" {
		return nil, "", nil
	}
	fi := &application.FulfillmentInput{PickupPoint: pickupPoint}
	if supplierID != nil {
		id, err := uuid.Parse(*supplierID)
		if err != nil {
			return nil, "invalid supplier_id", err
		}
		fi.SupplierID = &id
	}
	if warehouseBinID != nil {
		id, err := uuid.Parse(*warehouseBinID)
		if err != nil {
			return nil, "invalid warehouse_bin_id", err
		}
		fi.WarehouseBinID = &id
	}
	return fi, "", nil
}

// --- Private helpers ---

func toOrderResponse(o *domain.Order) dto.OrderResponse {
	r := dto.OrderResponse{
		ID:               o.ID.String(),
		UserID:           o.UserID.String(),
		ItemID:           o.ItemID.String(),
		Quantity:         o.Quantity,
		UnitPrice:        o.UnitPrice,
		TotalAmount:      o.TotalAmount,
		Status:           string(o.Status),
		SettlementMarker: o.SettlementMarker,
		Notes:            o.Notes,
		AutoCloseAt:      o.AutoCloseAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:        o.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        o.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
	if o.CampaignID != nil {
		s := o.CampaignID.String()
		r.CampaignID = &s
	}
	if o.PaidAt != nil {
		s := o.PaidAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		r.PaidAt = &s
	}
	if o.CancelledAt != nil {
		s := o.CancelledAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		r.CancelledAt = &s
	}
	if o.RefundedAt != nil {
		s := o.RefundedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		r.RefundedAt = &s
	}
	return r
}
