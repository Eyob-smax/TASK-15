package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fitcommerce/internal/application"
	"fitcommerce/internal/http/dto"
)

// DashboardHandler handles HTTP requests for the dashboard KPI endpoint.
type DashboardHandler struct {
	svc application.DashboardService
}

// NewDashboardHandler creates a DashboardHandler backed by the given service.
func NewDashboardHandler(svc application.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// GetKPIs handles GET /api/v1/dashboard/kpis.
// Optional query params: location_id, period, coach_id, category, from, to.
func (h *DashboardHandler) GetKPIs(c echo.Context) error {
	var locationID *uuid.UUID
	if v := c.QueryParam("location_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid location_id"))
		}
		locationID = &id
	}

	period := c.QueryParam("period")
	if period == "" {
		period = "monthly"
	}

	var coachID *uuid.UUID
	if v := c.QueryParam("coach_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid coach_id"))
		}
		coachID = &id
	}

	category := c.QueryParam("category")
	from := c.QueryParam("from")
	to := c.QueryParam("to")

	kpis, err := h.svc.GetKPIs(c.Request().Context(), locationID, period, coachID, category, from, to)
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := dto.KPIDashboardResponse{
		MemberGrowth:      toKPIMetric(kpis.MemberGrowth),
		Churn:             toKPIMetric(kpis.Churn),
		RenewalRate:       toKPIMetric(kpis.RenewalRate),
		Engagement:        toKPIMetric(kpis.Engagement),
		ClassFillRate:     toKPIMetric(kpis.ClassFillRate),
		CoachProductivity: toKPIMetric(kpis.CoachProductivity),
	}
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: resp})
}

func toKPIMetric(v application.KPIValue) dto.KPIMetric {
	return dto.KPIMetric{
		Value:         v.Value,
		PreviousValue: v.PreviousValue,
		ChangePercent: v.ChangePercent,
		Period:        v.Period,
	}
}
