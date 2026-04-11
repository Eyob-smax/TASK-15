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
