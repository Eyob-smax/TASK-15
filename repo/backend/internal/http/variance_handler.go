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

// VarianceHandler handles HTTP requests for variance endpoints.
type VarianceHandler struct {
	svc application.VarianceService
}

// NewVarianceHandler creates a VarianceHandler backed by the given service.
func NewVarianceHandler(svc application.VarianceService) *VarianceHandler {
	return &VarianceHandler{svc: svc}
}

// ListVariances handles GET /api/v1/variances.
func (h *VarianceHandler) ListVariances(c echo.Context) error {
	var statusFilter *domain.VarianceStatus
	if s := c.QueryParam("status"); s != "" {
		v := domain.VarianceStatus(s)
		statusFilter = &v
	}

	page, pageSize := parsePagination(c)

	variances, total, err := h.svc.List(c.Request().Context(), statusFilter, page, pageSize)
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.VarianceResponse, len(variances))
	for i := range variances {
		resp[i] = toVarianceResponse(&variances[i])
	}

	return c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       resp,
		Pagination: paginationMeta(page, pageSize, total),
	})
}

// GetVariance handles GET /api/v1/variances/:id.
func (h *VarianceHandler) GetVariance(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid variance id"))
	}

	variance, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: toVarianceResponse(variance)})
}

// ResolveVariance handles POST /api/v1/variances/:id/resolve.
func (h *VarianceHandler) ResolveVariance(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid variance id"))
	}

	var req dto.ResolveVarianceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", err.Error()))
	}
	if req.Action == "adjustment" && req.QuantityChange == nil {
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", "quantity_change is required for adjustment resolutions"))
	}
	if req.Action == "return" && req.QuantityChange != nil {
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", "quantity_change is only valid for adjustment resolutions"))
	}

	if err := h.svc.Resolve(c.Request().Context(), id, req.Action, req.ResolutionNotes, req.QuantityChange, user.ID); err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: map[string]string{"message": "resolved"}})
}

// --- Private helpers ---

func toVarianceResponse(v *domain.VarianceRecord) dto.VarianceResponse {
	var resolvedAt *string
	if v.ResolvedAt != nil {
		s := v.ResolvedAt.UTC().Format(time.RFC3339)
		resolvedAt = &s
	}

	return dto.VarianceResponse{
		ID:                 v.ID.String(),
		POLineID:           v.POLineID.String(),
		Type:               string(v.Type),
		ExpectedValue:      v.ExpectedValue,
		ActualValue:        v.ActualValue,
		DifferenceAmount:   v.DifferenceAmount,
		Status:             string(v.Status),
		ResolutionDueDate:  v.ResolutionDueDate.UTC().Format(time.RFC3339),
		ResolvedAt:         resolvedAt,
		ResolutionAction:   v.ResolutionAction,
		ResolutionNotes:    v.ResolutionNotes,
		QuantityChange:     v.QuantityChange,
		RequiresEscalation: v.RequiresEscalation(),
		IsOverdue:          v.IsOverdue(time.Now()),
		CreatedAt:          v.CreatedAt.UTC().Format(time.RFC3339),
	}
}
