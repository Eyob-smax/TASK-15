package api_tests

import (
	"net/http"
	"testing"
	"time"
)

func TestDashboard_MemberIsForbidden(t *testing.T) {
	app := newIntegrationApp(t)
	member := app.seedUser(t, "member", nil)

	rec := app.get(t, "/api/v1/dashboard/kpis", app.login(t, member))
	requireStatus(t, rec, http.StatusForbidden)
}

func TestDashboard_CoachWithoutLocationIsForbidden(t *testing.T) {
	app := newIntegrationApp(t)
	coach := app.seedUser(t, "coach", nil) // nil = no location assigned

	rec := app.get(t, "/api/v1/dashboard/kpis", app.login(t, coach))
	requireStatus(t, rec, http.StatusForbidden)

	if decodeError(t, rec).Error.Code != "FORBIDDEN" {
		t.Fatal("expected FORBIDDEN for coach without assigned location")
	}
}

func TestDashboard_InvalidDateFiltersReturn422(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	cookies := app.login(t, admin)

	badFrom := app.get(t, "/api/v1/dashboard/kpis?from=not-a-date", cookies)
	requireStatus(t, badFrom, http.StatusUnprocessableEntity)
	if decodeError(t, badFrom).Error.Code != "VALIDATION_ERROR" {
		t.Fatal("expected VALIDATION_ERROR for invalid from date")
	}

	badTo := app.get(t, "/api/v1/dashboard/kpis?to=31-01-2026", cookies)
	requireStatus(t, badTo, http.StatusUnprocessableEntity)
	if decodeError(t, badTo).Error.Code != "VALIDATION_ERROR" {
		t.Fatal("expected VALIDATION_ERROR for invalid to date")
	}
}

func TestDashboard_AdminAndCoachCanReadKPIs(t *testing.T) {
	app := newIntegrationApp(t)
	location := app.seedLocation(t, "Downtown Club")
	admin := app.seedUser(t, "administrator", nil)
	coachUser := app.seedUser(t, "coach", &location.ID)
	coachID := app.seedCoachRecord(t, coachUser.ID, location.ID, "Strength")
	memberUser := app.seedUser(t, "member", nil)
	app.seedMemberRecord(t, memberUser.ID, location.ID, "active", time.Now().UTC().AddDate(0, 0, -3), time.Now().UTC())

	query := "/api/v1/dashboard/kpis?location_id=" + location.ID.String() +
		"&coach_id=" + coachID.String() +
		"&period=monthly&from=2026-01-01&to=2026-01-31"

	adminRec := app.get(t, query, app.login(t, admin))
	requireStatus(t, adminRec, http.StatusOK)

	adminBody := decodeSuccess[map[string]any](t, adminRec)
	if adminBody["member_growth"] == nil || adminBody["coach_productivity"] == nil {
		t.Fatalf("expected KPI payload, got %#v", adminBody)
	}

	coachRec := app.get(t, query, app.login(t, coachUser))
	requireStatus(t, coachRec, http.StatusOK)
}
