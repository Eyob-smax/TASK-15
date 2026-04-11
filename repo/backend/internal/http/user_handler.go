package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
	"fitcommerce/internal/http/dto"
)

// UserHandler handles HTTP requests for user management endpoints.
type UserHandler struct {
	svc application.UserService
}

// NewUserHandler creates a UserHandler backed by the given service.
func NewUserHandler(svc application.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// CreateUser handles POST /admin/users.
func (h *UserHandler) CreateUser(c echo.Context) error {
	var req dto.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", err.Error()))
	}

	var locationID *uuid.UUID
	if req.LocationID != nil {
		id, err := uuid.Parse(*req.LocationID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid location_id"))
		}
		locationID = &id
	}

	user, err := h.svc.Create(c.Request().Context(), req.Email, req.Password, domain.UserRole(req.Role), req.DisplayName, locationID)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusCreated, dto.SuccessResponse{Data: toUserResponse(user)})
}

// ListUsers handles GET /admin/users.
func (h *UserHandler) ListUsers(c echo.Context) error {
	page, pageSize := parsePagination(c)

	users, total, err := h.svc.List(c.Request().Context(), page, pageSize)
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.UserResponse, len(users))
	for i := range users {
		resp[i] = toUserResponse(&users[i])
	}
	return c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       resp,
		Pagination: paginationMeta(page, pageSize, total),
	})
}

// GetUser handles GET /admin/users/:id.
func (h *UserHandler) GetUser(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid user id"))
	}

	user, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: toUserResponse(user)})
}

// UpdateUser handles PUT /admin/users/:id.
func (h *UserHandler) UpdateUser(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid user id"))
	}

	var req dto.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", err.Error()))
	}

	existing, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return HandleDomainError(c, err)
	}

	if req.DisplayName != "" {
		existing.DisplayName = req.DisplayName
	}
	if req.Role != "" {
		existing.Role = domain.UserRole(req.Role)
	}
	if req.LocationID != nil {
		lid, err := uuid.Parse(*req.LocationID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid location_id"))
		}
		existing.LocationID = &lid
	}

	if err := h.svc.Update(c.Request().Context(), existing); err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: toUserResponse(existing)})
}

// DeactivateUser handles POST /admin/users/:id/deactivate.
func (h *UserHandler) DeactivateUser(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid user id"))
	}

	if err := h.svc.Deactivate(c.Request().Context(), id); err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: map[string]string{"message": "user deactivated"}})
}

// --- Private helpers ---

func toUserResponse(u *domain.User) dto.UserResponse {
	resp := dto.UserResponse{
		ID:          u.ID.String(),
		Email:       u.Email,
		Role:        string(u.Role),
		Status:      string(u.Status),
		DisplayName: u.DisplayName,
		CreatedAt:   u.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   u.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
	if u.LocationID != nil {
		s := u.LocationID.String()
		resp.LocationID = &s
	}
	return resp
}
