package http

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
	"fitcommerce/internal/http/dto"
	"fitcommerce/internal/security"
)

// ProcurementHandler handles HTTP requests for purchase order endpoints.
type ProcurementHandler struct {
	svc application.PurchaseOrderService
}

// NewProcurementHandler creates a ProcurementHandler backed by the given service.
func NewProcurementHandler(svc application.PurchaseOrderService) *ProcurementHandler {
	return &ProcurementHandler{svc: svc}
}

// CreatePO handles POST /api/v1/purchase-orders.
func (h *ProcurementHandler) CreatePO(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	var req dto.CreatePurchaseOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", err.Error()))
	}

	supplierID, err := uuid.Parse(req.SupplierID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid supplier_id"))
	}

	po := domain.PurchaseOrder{
		SupplierID: supplierID,
		CreatedBy:  user.ID,
	}

	lines := make([]domain.PurchaseOrderLine, len(req.Lines))
	for i, l := range req.Lines {
		itemID, err := uuid.Parse(l.ItemID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid item_id: "+l.ItemID))
		}
		lines[i] = domain.PurchaseOrderLine{
			ItemID:           itemID,
			OrderedQuantity:  l.OrderedQuantity,
			OrderedUnitPrice: l.OrderedUnitPrice,
		}
	}

	created, err := h.svc.Create(c.Request().Context(), &po, lines)
	if err != nil {
		return HandleDomainError(c, err)
	}

	createdPO, createdLines, err := h.svc.Get(c.Request().Context(), created.ID)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusCreated, dto.SuccessResponse{Data: toPOResponse(createdPO, createdLines)})
}

// ListPOs handles GET /api/v1/purchase-orders.
func (h *ProcurementHandler) ListPOs(c echo.Context) error {
	page, pageSize := parsePagination(c)

	pos, total, err := h.svc.List(c.Request().Context(), page, pageSize)
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.PurchaseOrderResponse, len(pos))
	for i := range pos {
		resp[i] = toPOResponse(&pos[i], nil)
	}

	return c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       resp,
		Pagination: paginationMeta(page, pageSize, total),
	})
}

// GetPO handles GET /api/v1/purchase-orders/:id.
func (h *ProcurementHandler) GetPO(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid purchase order id"))
	}

	po, lines, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: toPOResponse(po, lines)})
}

// ApprovePO handles POST /api/v1/purchase-orders/:id/approve.
func (h *ProcurementHandler) ApprovePO(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid purchase order id"))
	}

	if err := h.svc.Approve(c.Request().Context(), id, user.ID); err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: map[string]string{"message": "approved"}})
}

// ReceivePO handles POST /api/v1/purchase-orders/:id/receive.
func (h *ProcurementHandler) ReceivePO(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid purchase order id"))
	}

	var req dto.ReceivePurchaseOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", err.Error()))
	}

	receivedLines := make([]application.ReceivedLineInput, len(req.Lines))
	for i, l := range req.Lines {
		poLineID, err := uuid.Parse(l.POLineID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid po_line_id: "+l.POLineID))
		}
		receivedLines[i] = application.ReceivedLineInput{
			POLineID:          poLineID,
			ReceivedQuantity:  l.ReceivedQuantity,
			ReceivedUnitPrice: l.ReceivedUnitPrice,
		}
	}

	if err := h.svc.Receive(c.Request().Context(), id, receivedLines, user.ID); err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: map[string]string{"message": "received"}})
}

// ReturnPO handles POST /api/v1/purchase-orders/:id/return.
func (h *ProcurementHandler) ReturnPO(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid purchase order id"))
	}

	if err := h.svc.Return(c.Request().Context(), id, user.ID); err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: map[string]string{"message": "returned"}})
}

// VoidPO handles POST /api/v1/purchase-orders/:id/void.
func (h *ProcurementHandler) VoidPO(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid purchase order id"))
	}

	if err := h.svc.Void(c.Request().Context(), id, user.ID); err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: map[string]string{"message": "voided"}})
}

// --- Private helpers ---

func toPOResponse(po *domain.PurchaseOrder, lines []domain.PurchaseOrderLine) dto.PurchaseOrderResponse {
	lineResponses := make([]dto.POLineResponse, len(lines))
	for i := range lines {
		lineResponses[i] = toPOLineResponse(&lines[i])
	}

	return dto.PurchaseOrderResponse{
		ID:          po.ID.String(),
		SupplierID:  po.SupplierID.String(),
		Status:      string(po.Status),
		TotalAmount: po.TotalAmount,
		CreatedBy:   po.CreatedBy.String(),
		ApprovedBy:  formatOptUUID(po.ApprovedBy),
		CreatedAt:   po.CreatedAt.UTC().Format(time.RFC3339),
		ApprovedAt:  formatOptTime(po.ApprovedAt),
		ReceivedAt:  formatOptTime(po.ReceivedAt),
		Version:     po.Version,
		Lines:       lineResponses,
	}
}

func toPOLineResponse(line *domain.PurchaseOrderLine) dto.POLineResponse {
	return dto.POLineResponse{
		ID:                line.ID.String(),
		PurchaseOrderID:   line.PurchaseOrderID.String(),
		ItemID:            line.ItemID.String(),
		OrderedQuantity:   line.OrderedQuantity,
		OrderedUnitPrice:  line.OrderedUnitPrice,
		ReceivedQuantity:  line.ReceivedQuantity,
		ReceivedUnitPrice: line.ReceivedUnitPrice,
	}
}

func formatOptUUID(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

func formatOptTime(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}
