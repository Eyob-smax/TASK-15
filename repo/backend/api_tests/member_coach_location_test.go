package api_tests

import (
	"net/http"
	"testing"
	"time"
)

func TestPersonnel_OperationsManagerIsScopedToAssignedLocation(t *testing.T) {
	app := newIntegrationApp(t)
	locationA := app.seedLocation(t, "North Club")
	locationB := app.seedLocation(t, "South Club")
	ops := app.seedUser(t, "operations_manager", &locationA.ID)
	opsCookies := app.login(t, ops)

	memberAUser := app.seedUser(t, "member", nil)
	memberBUser := app.seedUser(t, "member", nil)
	memberAID := app.seedMemberRecord(t, memberAUser.ID, locationA.ID, "active", time.Now().UTC().AddDate(0, 0, -2), time.Now().UTC())
	memberBID := app.seedMemberRecord(t, memberBUser.ID, locationB.ID, "active", time.Now().UTC().AddDate(0, 0, -2), time.Now().UTC())

	locationsRec := app.get(t, "/api/v1/locations", opsCookies)
	requireStatus(t, locationsRec, http.StatusOK)
	locations, _ := decodePaginated[map[string]any](t, locationsRec)
	if len(locations) != 1 || locations[0]["id"] != locationA.ID.String() {
		t.Fatalf("expected ops manager to see only own location, got %#v", locations)
	}

	memberListRec := app.get(t, "/api/v1/members?location_id="+locationB.ID.String(), opsCookies)
	requireStatus(t, memberListRec, http.StatusOK)
	members, _ := decodePaginated[map[string]any](t, memberListRec)
	if len(members) != 1 || members[0]["id"] != memberAID.String() {
		t.Fatalf("expected member list to be forced to ops manager location, got %#v", members)
	}

	forbiddenMemberRec := app.get(t, "/api/v1/members/"+memberBID.String(), opsCookies)
	requireStatus(t, forbiddenMemberRec, http.StatusForbidden)
}

func TestPersonnel_AdminCanAccessCrossLocationMemberAndCoachRecords(t *testing.T) {
	app := newIntegrationApp(t)
	locationA := app.seedLocation(t, "East Club")
	locationB := app.seedLocation(t, "West Club")
	admin := app.seedUser(t, "administrator", nil)
	adminCookies := app.login(t, admin)

	coachUser := app.seedUser(t, "coach", &locationB.ID)
	coachID := app.seedCoachRecord(t, coachUser.ID, locationB.ID, "Conditioning")
	memberUser := app.seedUser(t, "member", nil)
	memberID := app.seedMemberRecord(t, memberUser.ID, locationA.ID, "active", time.Now().UTC().AddDate(0, 0, -1), time.Now().UTC())

	memberRec := app.get(t, "/api/v1/members/"+memberID.String(), adminCookies)
	requireStatus(t, memberRec, http.StatusOK)

	coachRec := app.get(t, "/api/v1/coaches/"+coachID.String(), adminCookies)
	requireStatus(t, coachRec, http.StatusOK)

	coachListRec := app.get(t, "/api/v1/coaches?location_id="+locationB.ID.String(), adminCookies)
	requireStatus(t, coachListRec, http.StatusOK)
	coaches, _ := decodePaginated[map[string]any](t, coachListRec)
	if len(coaches) != 1 {
		t.Fatalf("expected one coach for the filtered location, got %#v", coaches)
	}
}
