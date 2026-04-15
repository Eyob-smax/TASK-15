package api_tests

import (
	"net/http"
	"testing"
	"time"
)

func TestPersonnel_AdminCanCreateLocationMemberAndCoach(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	adminCookies := app.login(t, admin)

	createLocationRec := app.post(t, "/api/v1/locations", map[string]any{
		"name":     "Central Club",
		"address":  "77 Core Road",
		"timezone": "UTC",
	}, adminCookies)
	requireStatus(t, createLocationRec, http.StatusCreated)
	location := decodeSuccess[map[string]any](t, createLocationRec)

	memberUser := app.seedUser(t, "member", nil)
	createMemberRec := app.post(t, "/api/v1/members", map[string]any{
		"user_id":     memberUser.ID.String(),
		"location_id": location["id"].(string),
	}, adminCookies)
	requireStatus(t, createMemberRec, http.StatusCreated)
	member := decodeSuccess[map[string]any](t, createMemberRec)

	coachUser := app.seedUser(t, "coach", nil)
	createCoachRec := app.post(t, "/api/v1/coaches", map[string]any{
		"user_id":        coachUser.ID.String(),
		"location_id":    location["id"].(string),
		"specialization": "Mobility",
	}, adminCookies)
	requireStatus(t, createCoachRec, http.StatusCreated)
	coach := decodeSuccess[map[string]any](t, createCoachRec)

	memberGetRec := app.get(t, "/api/v1/members/"+member["id"].(string), adminCookies)
	requireStatus(t, memberGetRec, http.StatusOK)

	coachGetRec := app.get(t, "/api/v1/coaches/"+coach["id"].(string), adminCookies)
	requireStatus(t, coachGetRec, http.StatusOK)

	locationsRec := app.get(t, "/api/v1/locations", adminCookies)
	requireStatus(t, locationsRec, http.StatusOK)
	locs, _ := decodePaginated[map[string]any](t, locationsRec)
	if len(locs) == 0 {
		t.Fatal("expected at least one location in admin list")
	}
}

func TestPersonnel_BadIDsAndValidationReturnExpectedStatus(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	adminCookies := app.login(t, admin)

	badLocationCreate := app.post(t, "/api/v1/locations", map[string]any{}, adminCookies)
	requireStatus(t, badLocationCreate, http.StatusUnprocessableEntity)

	badMemberCreate := app.post(t, "/api/v1/members", map[string]any{
		"user_id":     "not-a-uuid",
		"location_id": "not-a-uuid",
	}, adminCookies)
	requireStatus(t, badMemberCreate, http.StatusBadRequest)

	badCoachCreate := app.post(t, "/api/v1/coaches", map[string]any{
		"user_id":     "not-a-uuid",
		"location_id": "not-a-uuid",
	}, adminCookies)
	requireStatus(t, badCoachCreate, http.StatusBadRequest)

	badMemberList := app.get(t, "/api/v1/members?location_id=bad-uuid", adminCookies)
	requireStatus(t, badMemberList, http.StatusBadRequest)

	badCoachList := app.get(t, "/api/v1/coaches?location_id=bad-uuid", adminCookies)
	requireStatus(t, badCoachList, http.StatusBadRequest)

	badLocationGet := app.get(t, "/api/v1/locations/not-a-uuid", adminCookies)
	requireStatus(t, badLocationGet, http.StatusBadRequest)

	badMemberGet := app.get(t, "/api/v1/members/not-a-uuid", adminCookies)
	requireStatus(t, badMemberGet, http.StatusBadRequest)

	badCoachGet := app.get(t, "/api/v1/coaches/not-a-uuid", adminCookies)
	requireStatus(t, badCoachGet, http.StatusBadRequest)
}

func TestPersonnel_LocationGetIsForbiddenForOtherAssignedLocation(t *testing.T) {
	app := newIntegrationApp(t)
	locationA := app.seedLocation(t, "Ops Base")
	locationB := app.seedLocation(t, "Other Base")
	ops := app.seedUser(t, "operations_manager", &locationA.ID)

	forbiddenRec := app.get(t, "/api/v1/locations/"+locationB.ID.String(), app.login(t, ops))
	requireStatus(t, forbiddenRec, http.StatusForbidden)
}

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
