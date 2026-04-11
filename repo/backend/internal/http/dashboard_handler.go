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
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	var locationID *uuid.UUID
	if v := c.QueryParam("location_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid location_id"))
		}
		locationID = &id
	}

	// Coaches are scoped to their assigned location. A coach with no assigned
	// location cannot be safely scoped, so deny access entirely.
	if user.Role == domain.UserRoleCoach {
		if user.LocationID == nil {
			return c.JSON(http.StatusForbidden, NewErrorResponse("FORBIDDEN", "coach account has no assigned location"))
		}
		locationID = user.LocationID
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

	const dateLayout = "2006-01-02"
	if from != "" {
		if _, err := time.Parse(dateLayout, from); err != nil {
			return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", "from must be a date in YYYY-MM-DD format"))
		}
	}
	if to != "" {
		if _, err := time.Parse(dateLayout, to); err != nil {
			return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", "to must be a date in YYYY-MM-DD format"))
		}
	}

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
