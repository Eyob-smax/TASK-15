package http

import (
	"mime"
	"net/http"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
	"fitcommerce/internal/http/dto"
	"fitcommerce/internal/security"
)

// ReportHandler handles HTTP requests for report and export endpoints.
type ReportHandler struct {
	svc application.ReportService
}

// NewReportHandler creates a ReportHandler backed by the given service.
func NewReportHandler(svc application.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

// ListReports handles GET /reports.
func (h *ReportHandler) ListReports(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	reports, err := h.svc.List(c.Request().Context(), user.Role)
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.ReportResponse, len(reports))
	for i := range reports {
		resp[i] = toReportResponse(&reports[i])
	}
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: resp})
}

// GetReportData handles GET /reports/:id/data.
func (h *ReportHandler) GetReportData(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid report id"))
	}

	// Handler-layer role check: verify access before querying data.
	report, err := h.svc.GetReport(c.Request().Context(), id)
	if err != nil {
		return HandleDomainError(c, err)
	}
	allowed := false
	for _, r := range report.AllowedRoles {
		if r == user.Role {
			allowed = true
			break
		}
	}
	if !allowed {
		return c.JSON(http.StatusForbidden, NewErrorResponse("FORBIDDEN", "not authorized to access this report"))
	}

	filters := make(map[string]string)
	for k, v := range c.QueryParams() {
		if len(v) > 0 {
			filters[k] = v[0]
		}
	}

	// Coaches are scoped to their assigned location. A coach with no assigned
	// location cannot be safely scoped, so deny access entirely.
	if user.Role == domain.UserRoleCoach {
		if user.LocationID == nil {
			return c.JSON(http.StatusForbidden, NewErrorResponse("FORBIDDEN", "coach account has no assigned location"))
		}
		filters["location_id"] = user.LocationID.String()
	}

	// Service enforces role check as a second layer.
	result, err := h.svc.GetData(c.Request().Context(), id, filters, user.Role)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: result})
}

// RunExport handles POST /exports.
func (h *ReportHandler) RunExport(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	var req dto.CreateExportRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", err.Error()))
	}

	reportID, err := uuid.Parse(req.ReportID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid report_id"))
	}

	// Coaches are scoped to their assigned location — same enforcement as GetReportData.
	if user.Role == domain.UserRoleCoach {
		if user.LocationID == nil {
			return c.JSON(http.StatusForbidden, NewErrorResponse("FORBIDDEN", "coach account has no assigned location"))
		}
		if req.Parameters == nil {
			req.Parameters = make(map[string]string)
		}
		req.Parameters["location_id"] = user.LocationID.String()
	}

	job, err := h.svc.GenerateExport(c.Request().Context(), reportID, domain.ExportFormat(req.Format), req.Parameters, user.ID, user.Role)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusAccepted, dto.SuccessResponse{Data: toExportResponse(job)})
}

// GetExport handles GET /exports/:id.
func (h *ReportHandler) GetExport(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid export id"))
	}

	// Service enforces creator-or-admin access.
	job, err := h.svc.GetExport(c.Request().Context(), id, user.ID, user.Role)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: toExportResponse(job)})
}

// DownloadExport handles GET /exports/:id/download.
func (h *ReportHandler) DownloadExport(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid export id"))
	}

	// Service enforces creator-or-admin access.
	job, err := h.svc.DownloadExport(c.Request().Context(), id, user.ID, user.Role)
	if err != nil {
		return HandleDomainError(c, err)
	}

	if job.FilePath == "" {
		return c.JSON(http.StatusNotFound, NewErrorResponse("NOT_FOUND", "export file not found"))
	}

	filename := job.Filename
	if filename == "" {
		filename = filepath.Base(job.FilePath)
	}
	contentDisposition := mime.FormatMediaType("attachment", map[string]string{"filename": filename})
	if contentDisposition != "" {
		c.Response().Header().Set("Content-Disposition", contentDisposition)
	}
	if contentType := mime.TypeByExtension(filepath.Ext(filename)); contentType != "" {
		c.Response().Header().Set("Content-Type", contentType)
	} else {
		c.Response().Header().Set("Content-Type", "application/octet-stream")
	}
	c.Response().Header().Set("Cache-Control", "no-store")

	return c.File(job.FilePath)
}

// --- Private helpers ---

func toReportResponse(r *domain.ReportDefinition) dto.ReportResponse {
	roles := make([]string, len(r.AllowedRoles))
	for i, role := range r.AllowedRoles {
		roles[i] = string(role)
	}
	return dto.ReportResponse{
		ID:           r.ID.String(),
		Name:         r.Name,
		ReportType:   r.ReportType,
		Description:  r.Description,
		AllowedRoles: roles,
		CreatedAt:    r.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toExportResponse(j *domain.ExportJob) dto.ExportResponse {
	resp := dto.ExportResponse{
		ID:        j.ID.String(),
		ReportID:  j.ReportID.String(),
		Format:    string(j.Format),
		Filename:  j.Filename,
		Status:    string(j.Status),
		CreatedBy: j.CreatedBy.String(),
		CreatedAt: j.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
	if j.CompletedAt != nil {
		s := j.CompletedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		resp.CompletedAt = &s
	}
	return resp
}
