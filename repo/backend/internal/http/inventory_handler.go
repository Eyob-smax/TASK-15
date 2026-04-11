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

// InventoryHandler handles HTTP requests for the inventory and warehouse bin endpoints.
type InventoryHandler struct {
	svc application.InventoryService
}

// NewInventoryHandler creates an InventoryHandler backed by the given service.
func NewInventoryHandler(svc application.InventoryService) *InventoryHandler {
	return &InventoryHandler{svc: svc}
}

// GetSnapshots handles GET /api/v1/inventory/snapshots.
func (h *InventoryHandler) GetSnapshots(c echo.Context) error {
	var itemID *uuid.UUID
	if v := c.QueryParam("item_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid item_id"))
		}
		itemID = &id
	}
	var locationID *uuid.UUID
	if v := c.QueryParam("location_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid location_id"))
		}
		locationID = &id
	}

	snapshots, err := h.svc.GetSnapshots(c.Request().Context(), itemID, locationID)
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.InventorySnapshotResponse, len(snapshots))
	for i, sn := range snapshots {
		resp[i] = toSnapshotResponse(sn)
	}
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: resp})
}

// CreateAdjustment handles POST /api/v1/inventory/adjustments.
func (h *InventoryHandler) CreateAdjustment(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	var req dto.InventoryAdjustmentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}

	itemID, err := uuid.Parse(req.ItemID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid item_id"))
	}

	adj := &domain.InventoryAdjustment{
		ItemID:         itemID,
		QuantityChange: req.QuantityChange,
		Reason:         req.Reason,
		CreatedBy:      user.ID,
	}

	created, err := h.svc.CreateAdjustment(c.Request().Context(), adj)
	if err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusCreated, dto.SuccessResponse{Data: toAdjustmentResponse(*created)})
}

// ListAdjustments handles GET /api/v1/inventory/adjustments.
func (h *InventoryHandler) ListAdjustments(c echo.Context) error {
	page, pageSize := parsePagination(c)
	var itemID *uuid.UUID
	if v := c.QueryParam("item_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid item_id"))
		}
		itemID = &id
	}

	adjs, total, err := h.svc.ListAdjustments(c.Request().Context(), itemID, page, pageSize)
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.InventoryAdjustmentResponse, len(adjs))
	for i, a := range adjs {
		resp[i] = toAdjustmentResponse(a)
	}
	return c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       resp,
		Pagination: paginationMeta(page, pageSize, total),
	})
}

// CreateWarehouseBin handles POST /api/v1/warehouse-bins.
func (h *InventoryHandler) CreateWarehouseBin(c echo.Context) error {
	var req dto.CreateWarehouseBinRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}

	locationID, err := uuid.Parse(req.LocationID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid location_id"))
	}

	bin := &domain.WarehouseBin{
		LocationID:  locationID,
		Name:        req.Name,
		Description: req.Description,
	}

	created, err := h.svc.CreateWarehouseBin(c.Request().Context(), bin)
	if err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusCreated, dto.SuccessResponse{Data: toBinResponse(*created)})
}

// ListWarehouseBins handles GET /api/v1/warehouse-bins.
func (h *InventoryHandler) ListWarehouseBins(c echo.Context) error {
	page, pageSize := parsePagination(c)
	var locationID *uuid.UUID
	if v := c.QueryParam("location_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid location_id"))
		}
		locationID = &id
	}

	bins, total, err := h.svc.ListWarehouseBins(c.Request().Context(), locationID, page, pageSize)
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.WarehouseBinResponse, len(bins))
	for i, b := range bins {
		resp[i] = toBinResponse(b)
	}
	return c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       resp,
		Pagination: paginationMeta(page, pageSize, total),
	})
}

// GetWarehouseBin handles GET /api/v1/warehouse-bins/:id.
func (h *InventoryHandler) GetWarehouseBin(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid bin id"))
	}

	bin, err := h.svc.GetWarehouseBin(c.Request().Context(), id)
	if err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: toBinResponse(*bin)})
}

// --- Private helpers ---

func toSnapshotResponse(sn domain.InventorySnapshot) dto.InventorySnapshotResponse {
	r := dto.InventorySnapshotResponse{
		ID:             sn.ID.String(),
		ItemID:         sn.ItemID.String(),
		QuantityOnHand: sn.Quantity,
		SnapshotDate:   sn.RecordedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
	if sn.LocationID != nil {
		s := sn.LocationID.String()
		r.LocationID = &s
	}
	return r
}

func toAdjustmentResponse(a domain.InventoryAdjustment) dto.InventoryAdjustmentResponse {
	return dto.InventoryAdjustmentResponse{
		ID:             a.ID.String(),
		ItemID:         a.ItemID.String(),
		QuantityChange: a.QuantityChange,
		Reason:         a.Reason,
		CreatedBy:      a.CreatedBy.String(),
		CreatedAt:      a.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toBinResponse(b domain.WarehouseBin) dto.WarehouseBinResponse {
	return dto.WarehouseBinResponse{
		ID:          b.ID.String(),
		LocationID:  b.LocationID.String(),
		Name:        b.Name,
		Description: b.Description,
		CreatedAt:   b.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   b.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}
