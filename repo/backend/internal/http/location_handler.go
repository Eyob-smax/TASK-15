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

// LocationHandler handles HTTP requests for location endpoints.
type LocationHandler struct {
	svc application.LocationService
}

// NewLocationHandler creates a LocationHandler backed by the given service.
func NewLocationHandler(svc application.LocationService) *LocationHandler {
	return &LocationHandler{svc: svc}
}

// CreateLocation handles POST /api/v1/locations.
func (h *LocationHandler) CreateLocation(c echo.Context) error {
	var req dto.CreateLocationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", err.Error()))
	}

	loc := &domain.Location{
		Name:     req.Name,
		Address:  req.Address,
		Timezone: req.Timezone,
	}
	created, err := h.svc.Create(c.Request().Context(), loc)
	if err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusCreated, dto.SuccessResponse{Data: toLocationResponse(created)})
}

// ListLocations handles GET /api/v1/locations.
func (h *LocationHandler) ListLocations(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	page, pageSize := parsePagination(c)

	// Non-administrator users with a location assignment are scoped to their own
	// location only; they cannot enumerate records outside their assigned location.
	if user.Role != domain.UserRoleAdministrator && user.LocationID != nil {
		loc, err := h.svc.GetByID(c.Request().Context(), *user.LocationID)
		if err != nil {
			return HandleDomainError(c, err)
		}
		return c.JSON(http.StatusOK, dto.PaginatedResponse{
			Data:       []dto.LocationResponse{toLocationResponse(loc)},
			Pagination: paginationMeta(page, pageSize, 1),
		})
	}

	locs, total, err := h.svc.List(c.Request().Context(), page, pageSize)
	if err != nil {
		return HandleDomainError(c, err)
	}
	resp := make([]dto.LocationResponse, len(locs))
	for i := range locs {
		resp[i] = toLocationResponse(&locs[i])
	}
	return c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       resp,
		Pagination: paginationMeta(page, pageSize, total),
	})
}

// GetLocation handles GET /api/v1/locations/:id.
func (h *LocationHandler) GetLocation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid location id"))
	}

	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	// Non-admin users with a location assignment may only view their own location.
	if user.Role != domain.UserRoleAdministrator && user.LocationID != nil {
		if id != *user.LocationID {
			return c.JSON(http.StatusForbidden, NewErrorResponse("FORBIDDEN", "access denied for this location"))
		}
	}

	loc, err := h.svc.GetByID(c.Request().Context(), id)
	if err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: toLocationResponse(loc)})
}

func toLocationResponse(loc *domain.Location) dto.LocationResponse {
	return dto.LocationResponse{
		ID:        loc.ID.String(),
		Name:      loc.Name,
		Address:   loc.Address,
		Timezone:  loc.Timezone,
		IsActive:  loc.IsActive,
		CreatedAt: loc.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: loc.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
