package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fitcommerce/internal/application"
	"fitcommerce/internal/http/dto"
)

// LandedCostHandler handles HTTP requests for landed cost endpoints.
type LandedCostHandler struct {
	svc application.LandedCostService
}

// NewLandedCostHandler creates a LandedCostHandler backed by the given service.
func NewLandedCostHandler(svc application.LandedCostService) *LandedCostHandler {
	return &LandedCostHandler{svc: svc}
}

// GetLandedCosts handles GET /api/v1/procurement/landed-costs.
// Query params: item_id (UUID, required), period (string, e.g. "2026-Q1").
func (h *LandedCostHandler) GetLandedCosts(c echo.Context) error {
	itemIDStr := c.QueryParam("item_id")
	if itemIDStr == "" {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "item_id is required"))
	}
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid item_id"))
	}
	period := c.QueryParam("period")
	if period == "" {
		period = "current"
	}

	entries, err := h.svc.GetSummary(c.Request().Context(), itemID, period)
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.LandedCostResponse, len(entries))
	for i, e := range entries {
		resp[i] = dto.LandedCostResponse{
			ID:               e.ID.String(),
			ItemID:           e.ItemID.String(),
			PurchaseOrderID:  e.PurchaseOrderID.String(),
			POLineID:         e.POLineID.String(),
			Period:           e.Period,
			CostComponent:    e.CostComponent,
			RawAmount:        e.RawAmount,
			AllocatedAmount:  e.AllocatedAmount,
			AllocationMethod: e.AllocationMethod,
			CreatedAt:        e.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: resp})
}

// GetLandedCostsByPO handles GET /api/v1/procurement/landed-costs/:poId.
func (h *LandedCostHandler) GetLandedCostsByPO(c echo.Context) error {
	poID, err := uuid.Parse(c.Param("poId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid purchase order id"))
	}

	entries, err := h.svc.GetByPOID(c.Request().Context(), poID)
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.LandedCostResponse, len(entries))
	for i, e := range entries {
		resp[i] = dto.LandedCostResponse{
			ID:               e.ID.String(),
			ItemID:           e.ItemID.String(),
			PurchaseOrderID:  e.PurchaseOrderID.String(),
			POLineID:         e.POLineID.String(),
			Period:           e.Period,
			CostComponent:    e.CostComponent,
			RawAmount:        e.RawAmount,
			AllocatedAmount:  e.AllocatedAmount,
			AllocationMethod: e.AllocationMethod,
			CreatedAt:        e.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: resp})
}
