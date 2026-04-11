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

// CampaignHandler handles HTTP requests for the group-buy campaign endpoints.
type CampaignHandler struct {
	svc application.CampaignService
}

// NewCampaignHandler creates a CampaignHandler backed by the given service.
func NewCampaignHandler(svc application.CampaignService) *CampaignHandler {
	return &CampaignHandler{svc: svc}
}

// CreateCampaign handles POST /api/v1/campaigns.
func (h *CampaignHandler) CreateCampaign(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	var req dto.CreateCampaignRequest
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

	campaign := &domain.GroupBuyCampaign{
		ItemID:      itemID,
		MinQuantity: req.MinQuantity,
		MaxQuantity: req.MaxQuantity,
		CutoffTime:  req.CutoffTime,
		CreatedBy:   user.ID,
	}

	created, err := h.svc.Create(c.Request().Context(), campaign)
	if err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusCreated, dto.SuccessResponse{Data: toCampaignResponse(created)})
}

// ListCampaigns handles GET /api/v1/campaigns.
func (h *CampaignHandler) ListCampaigns(c echo.Context) error {
	page, pageSize := parsePagination(c)
	campaigns, total, err := h.svc.List(c.Request().Context(), page, pageSize)
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.CampaignResponse, len(campaigns))
	for i := range campaigns {
		resp[i] = toCampaignResponse(&campaigns[i])
	}
	return c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       resp,
		Pagination: paginationMeta(page, pageSize, total),
	})
}

// GetCampaign handles GET /api/v1/campaigns/:id.
func (h *CampaignHandler) GetCampaign(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid campaign id"))
	}

	campaign, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: toCampaignResponse(campaign)})
}

// JoinCampaign handles POST /api/v1/campaigns/:id/join.
func (h *CampaignHandler) JoinCampaign(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid campaign id"))
	}

	var req dto.JoinCampaignRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", err.Error()))
	}

	participant, err := h.svc.Join(c.Request().Context(), id, user.ID, req.Quantity)
	if err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusCreated, dto.SuccessResponse{Data: toParticipantResponse(participant)})
}

// CancelCampaign handles POST /api/v1/campaigns/:id/cancel.
func (h *CampaignHandler) CancelCampaign(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid campaign id"))
	}

	if err := h.svc.Cancel(c.Request().Context(), id, user.ID); err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: map[string]string{"message": "campaign cancelled"}})
}

// EvaluateCampaign handles POST /api/v1/campaigns/:id/evaluate.
func (h *CampaignHandler) EvaluateCampaign(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid campaign id"))
	}

	if err := h.svc.EvaluateAtCutoff(c.Request().Context(), id, time.Now().UTC(), user.ID); err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: map[string]string{"message": "campaign evaluated"}})
}

// --- Private helpers ---

func toCampaignResponse(c *domain.GroupBuyCampaign) dto.CampaignResponse {
	r := dto.CampaignResponse{
		ID:                  c.ID.String(),
		ItemID:              c.ItemID.String(),
		MinQuantity:         c.MinQuantity,
		MaxQuantity:         c.MaxQuantity,
		CurrentCommittedQty: c.CurrentCommittedQty,
		CutoffTime:          c.CutoffTime.UTC().Format("2006-01-02T15:04:05Z07:00"),
		Status:              string(c.Status),
		CreatedBy:           c.CreatedBy.String(),
		CreatedAt:           c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
	if c.EvaluatedAt != nil {
		s := c.EvaluatedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		r.EvaluatedAt = &s
	}
	return r
}

func toParticipantResponse(p *domain.GroupBuyParticipant) dto.ParticipantResponse {
	return dto.ParticipantResponse{
		ID:         p.ID.String(),
		CampaignID: p.CampaignID.String(),
		UserID:     p.UserID.String(),
		Quantity:   p.Quantity,
		OrderID:    p.OrderID.String(),
		JoinedAt:   p.JoinedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}
