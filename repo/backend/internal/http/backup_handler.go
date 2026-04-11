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

// BackupHandler handles HTTP requests for backup endpoints.
type BackupHandler struct {
	svc application.BackupService
}

// NewBackupHandler creates a BackupHandler backed by the given service.
func NewBackupHandler(svc application.BackupService) *BackupHandler {
	return &BackupHandler{svc: svc}
}

// TriggerBackup handles POST /admin/backups.
func (h *BackupHandler) TriggerBackup(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	callerID := user.ID
	run, err := h.svc.Trigger(c.Request().Context(), &callerID)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusAccepted, dto.SuccessResponse{Data: toBackupResponse(run)})
}

// ListBackups handles GET /admin/backups.
func (h *BackupHandler) ListBackups(c echo.Context) error {
	page, pageSize := parsePagination(c)

	runs, total, err := h.svc.List(c.Request().Context(), page, pageSize)
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.BackupResponse, len(runs))
	for i := range runs {
		resp[i] = toBackupResponse(&runs[i])
	}
	return c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       resp,
		Pagination: paginationMeta(page, pageSize, total),
	})
}

// VerifyBackup handles GET /admin/backups/:id/verify.
func (h *BackupHandler) VerifyBackup(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("INVALID_ID", "invalid backup ID"))
	}

	run, err := h.svc.VerifyIntegrity(c.Request().Context(), id)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: map[string]interface{}{
		"valid":     true,
		"backup_id": run.ID.String(),
		"checksum":  run.Checksum,
	}})
}

// --- Private helpers ---

func toBackupResponse(r *domain.BackupRun) dto.BackupResponse {
	resp := dto.BackupResponse{
		ID:                r.ID.String(),
		ArchivePath:       r.ArchivePath,
		Checksum:          r.Checksum,
		ChecksumAlgorithm: r.ChecksumAlgorithm,
		Status:            string(r.Status),
		FileSize:          r.FileSize,
		StartedAt:         r.StartedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
	if r.CompletedAt != nil {
		s := r.CompletedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		resp.CompletedAt = &s
	}
	return resp
}

