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

// CoachHandler handles HTTP requests for coach endpoints.
type CoachHandler struct {
	svc application.CoachService
}

// NewCoachHandler creates a CoachHandler backed by the given service.
func NewCoachHandler(svc application.CoachService) *CoachHandler {
	return &CoachHandler{svc: svc}
}

// CreateCoach handles POST /api/v1/coaches.
func (h *CoachHandler) CreateCoach(c echo.Context) error {
	var req struct {
		UserID         string `json:"user_id" validate:"required"`
		LocationID     string `json:"location_id" validate:"required"`
		Specialization string `json:"specialization"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", err.Error()))
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid user_id"))
	}
	locationID, err := uuid.Parse(req.LocationID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid location_id"))
	}

	coach := &domain.Coach{
		UserID:         userID,
		LocationID:     locationID,
		Specialization: req.Specialization,
	}
	created, err := h.svc.Create(c.Request().Context(), coach)
	if err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusCreated, dto.SuccessResponse{Data: toCoachResponse(created)})
}

// ListCoaches handles GET /api/v1/coaches.
func (h *CoachHandler) ListCoaches(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	page, pageSize := parsePagination(c)

	var locationID *uuid.UUID
	if v := c.QueryParam("location_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid location_id"))
		}
		locationID = &id
	}

	coaches, total, err := h.svc.ListForActor(c.Request().Context(), user, locationID, page, pageSize)
	if err != nil {
		return HandleDomainError(c, err)
	}
	resp := make([]dto.CoachResponse, len(coaches))
	for i := range coaches {
		resp[i] = toCoachResponse(&coaches[i])
	}
	return c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       resp,
		Pagination: paginationMeta(page, pageSize, total),
	})
}

// GetCoach handles GET /api/v1/coaches/:id.
func (h *CoachHandler) GetCoach(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid coach id"))
	}

	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	coach, err := h.svc.GetByIDForActor(c.Request().Context(), user, id)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: toCoachResponse(coach)})
}

func toCoachResponse(c *domain.Coach) dto.CoachResponse {
	return dto.CoachResponse{
		ID:             c.ID.String(),
		UserID:         c.UserID.String(),
		LocationID:     c.LocationID.String(),
		Specialization: c.Specialization,
		IsActive:       c.IsActive,
		CreatedAt:      c.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      c.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
