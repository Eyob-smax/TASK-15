package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
	"fitcommerce/internal/http/dto"
	"fitcommerce/internal/security"
)

// AdminHandler handles HTTP requests for admin/audit endpoints.
type AdminHandler struct {
	auditSvc application.AuditService
}

// NewAdminHandler creates an AdminHandler backed by the given audit service.
func NewAdminHandler(auditSvc application.AuditService) *AdminHandler {
	return &AdminHandler{auditSvc: auditSvc}
}

// GetAuditLog handles GET /admin/audit-log.
func (h *AdminHandler) GetAuditLog(c echo.Context) error {
	eventTypeFilter := c.QueryParam("event_type")
	entityTypeFilter := c.QueryParam("entity_type")
	page, pageSize := parsePagination(c)

	var events []domain.AuditEvent
	var total int
	var err error

	if eventTypeFilter != "" {
		// Filter by event_type column (e.g. "auth.login.success").
		events, total, err = h.auditSvc.ListByEventTypes(c.Request().Context(), []string{eventTypeFilter}, page, pageSize)
	} else {
		// Filter by entity_type column (or return all when empty).
		events, total, err = h.auditSvc.List(c.Request().Context(), entityTypeFilter, nil, page, pageSize)
	}

	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.AuditEventResponse, len(events))
	for i := range events {
		resp[i] = toAuditEventResponse(&events[i])
	}
	return c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       resp,
		Pagination: paginationMeta(page, pageSize, total),
	})
}

// GetSecurityEvents handles GET /admin/audit-log/security.
func (h *AdminHandler) GetSecurityEvents(c echo.Context) error {
	page, pageSize := parsePagination(c)

	securityEventTypes := []string{
		security.EventLoginFailure,
		security.EventLoginLockout,
		security.EventSessionExpired,
		security.EventCaptchaVerified,
		security.EventCaptchaFailed,
	}

	events, total, err := h.auditSvc.ListByEventTypes(c.Request().Context(), securityEventTypes, page, pageSize)
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.AuditEventResponse, len(events))
	for i := range events {
		resp[i] = toAuditEventResponse(&events[i])
	}
	return c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       resp,
		Pagination: paginationMeta(page, pageSize, total),
	})
}

// --- Private helpers ---

func toAuditEventResponse(e *domain.AuditEvent) dto.AuditEventResponse {
	return dto.AuditEventResponse{
		ID:            e.ID.String(),
		EventType:     e.EventType,
		EntityType:    e.EntityType,
		EntityID:      e.EntityID.String(),
		ActorID:       e.ActorID.String(),
		Details:       e.Details,
		IntegrityHash: e.IntegrityHash,
		CreatedAt:     e.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}
