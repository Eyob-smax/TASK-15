package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
	"fitcommerce/internal/http/dto"
)

// RetentionHandler handles HTTP requests for retention policy endpoints.
type RetentionHandler struct {
	svc application.RetentionService
}

// NewRetentionHandler creates a RetentionHandler backed by the given service.
func NewRetentionHandler(svc application.RetentionService) *RetentionHandler {
	return &RetentionHandler{svc: svc}
}

// ListRetentionPolicies handles GET /admin/retention-policies.
func (h *RetentionHandler) ListRetentionPolicies(c echo.Context) error {
	policies, err := h.svc.List(c.Request().Context())
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.RetentionPolicyResponse, len(policies))
	for i := range policies {
		resp[i] = toRetentionPolicyResponse(&policies[i])
	}
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: resp})
}

// GetRetentionPolicy handles GET /admin/retention-policies/:entity_type.
func (h *RetentionHandler) GetRetentionPolicy(c echo.Context) error {
	entityType := c.Param("entity_type")

	policy, err := h.svc.GetByEntityType(c.Request().Context(), entityType)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: toRetentionPolicyResponse(policy)})
}

// UpdateRetentionPolicy handles PUT /admin/retention-policies/:entity_type.
func (h *RetentionHandler) UpdateRetentionPolicy(c echo.Context) error {
	entityType := c.Param("entity_type")

	var req dto.UpdateRetentionPolicyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", err.Error()))
	}

	policy := domain.RetentionPolicy{
		EntityType:    entityType,
		RetentionDays: req.RetentionDays,
	}

	if err := h.svc.Update(c.Request().Context(), &policy); err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: toRetentionPolicyResponse(&policy)})
}

// --- Private helpers ---

func toRetentionPolicyResponse(p *domain.RetentionPolicy) dto.RetentionPolicyResponse {
	return dto.RetentionPolicyResponse{
		ID:            p.ID.String(),
		EntityType:    p.EntityType,
		RetentionDays: p.RetentionDays,
		UpdatedAt:     p.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}
