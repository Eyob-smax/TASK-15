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

// MemberHandler handles HTTP requests for member endpoints.
type MemberHandler struct {
	svc application.MemberService
}

// NewMemberHandler creates a MemberHandler backed by the given service.
func NewMemberHandler(svc application.MemberService) *MemberHandler {
	return &MemberHandler{svc: svc}
}

// CreateMember handles POST /api/v1/members.
func (h *MemberHandler) CreateMember(c echo.Context) error {
	var req struct {
		UserID     string `json:"user_id" validate:"required"`
		LocationID string `json:"location_id" validate:"required"`
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

	member := &domain.Member{UserID: userID, LocationID: locationID}
	created, err := h.svc.Create(c.Request().Context(), member)
	if err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusCreated, dto.SuccessResponse{Data: toMemberResponse(created)})
}

// ListMembers handles GET /api/v1/members.
func (h *MemberHandler) ListMembers(c echo.Context) error {
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

	members, total, err := h.svc.ListForActor(c.Request().Context(), user, locationID, page, pageSize)
	if err != nil {
		return HandleDomainError(c, err)
	}
	resp := make([]dto.MemberResponse, len(members))
	for i := range members {
		resp[i] = toMemberResponse(&members[i])
	}
	return c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       resp,
		Pagination: paginationMeta(page, pageSize, total),
	})
}

// GetMember handles GET /api/v1/members/:id.
func (h *MemberHandler) GetMember(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid member id"))
	}

	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	member, err := h.svc.GetByIDForActor(c.Request().Context(), user, id)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: toMemberResponse(member)})
}

func toMemberResponse(m *domain.Member) dto.MemberResponse {
	r := dto.MemberResponse{
		ID:               m.ID.String(),
		UserID:           m.UserID.String(),
		LocationID:       m.LocationID.String(),
		MembershipStatus: string(m.MembershipStatus),
		JoinedAt:         m.JoinedAt.UTC().Format(time.RFC3339),
		CreatedAt:        m.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:        m.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if m.RenewalDate != nil {
		s := m.RenewalDate.UTC().Format("2006-01-02")
		r.RenewalDate = &s
	}
	return r
}
