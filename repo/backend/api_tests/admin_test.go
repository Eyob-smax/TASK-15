package api_tests

import (
	"net/http"
	"testing"
)

func TestAdmin_CanCreateUserAndReadSecurityAudit(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	adminCookies := app.login(t, admin)

	createRec := app.post(t, "/api/v1/admin/users", map[string]any{
		"email":        "ops.audit@fitcommerce.test",
		"password":     defaultTestPassword,
		"role":         "operations_manager",
		"display_name": "Ops Audit",
	}, adminCookies)
	requireStatus(t, createRec, http.StatusCreated)

	_ = app.post(t, "/api/v1/auth/login", map[string]string{
		"email":    admin.Email,
		"password": "WrongPassword123!",
	}, nil)

	auditRec := app.get(t, "/api/v1/admin/audit-log/security", adminCookies)
	requireStatus(t, auditRec, http.StatusOK)

	events, pagination := decodePaginated[map[string]any](t, auditRec)
	if len(events) == 0 || pagination.TotalCount == 0 {
		t.Fatal("expected security audit endpoint to return at least one event")
	}

	allAuditRec := app.get(t, "/api/v1/admin/audit-log?entity_type=user", adminCookies)
	requireStatus(t, allAuditRec, http.StatusOK)
	allEvents, _ := decodePaginated[map[string]any](t, allAuditRec)
	if len(allEvents) == 0 {
		t.Fatal("expected audit-log endpoint to return at least one user event")
	}

	eventTypeAuditRec := app.get(t, "/api/v1/admin/audit-log?event_type=auth.login.success", adminCookies)
	requireStatus(t, eventTypeAuditRec, http.StatusOK)
	_, _ = decodePaginated[map[string]any](t, eventTypeAuditRec)
}

func TestAdmin_CanTriggerAndListBackups(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	adminCookies := app.login(t, admin)

	triggerRec := app.post(t, "/api/v1/admin/backups", map[string]any{}, adminCookies)
	requireStatus(t, triggerRec, http.StatusAccepted)

	backup := decodeSuccess[map[string]any](t, triggerRec)
	if backup["status"] != "completed" {
		t.Fatalf("expected completed backup status, got %#v", backup["status"])
	}

	listRec := app.get(t, "/api/v1/admin/backups", adminCookies)
	requireStatus(t, listRec, http.StatusOK)

	backups, pagination := decodePaginated[map[string]any](t, listRec)
	if len(backups) != 1 || pagination.TotalCount != 1 {
		t.Fatalf("expected exactly one backup run, got len=%d total=%d", len(backups), pagination.TotalCount)
	}

	verifyRec := app.get(t, "/api/v1/admin/backups/"+backup["id"].(string)+"/verify", adminCookies)
	requireStatus(t, verifyRec, http.StatusOK)
	verified := decodeSuccess[map[string]any](t, verifyRec)
	if verified["valid"] != true {
		t.Fatalf("expected backup verification to be valid, got %#v", verified)
	}
}

func TestAdmin_CanUpdateRetentionPolicyInDays(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	adminCookies := app.login(t, admin)

	updateRec := app.put(t, "/api/v1/admin/retention-policies/access_logs", map[string]any{
		"retention_days": 45,
	}, adminCookies)
	requireStatus(t, updateRec, http.StatusOK)

	updated := decodeSuccess[map[string]any](t, updateRec)
	if int(updated["retention_days"].(float64)) != 45 {
		t.Fatalf("expected retention_days 45, got %#v", updated["retention_days"])
	}
	if updated["updated_at"] == "" {
		t.Fatal("expected updated_at to be populated")
	}

	getRec := app.get(t, "/api/v1/admin/retention-policies/access_logs", adminCookies)
	requireStatus(t, getRec, http.StatusOK)

	policy := decodeSuccess[map[string]any](t, getRec)
	if int(policy["retention_days"].(float64)) != 45 {
		t.Fatalf("expected persisted retention_days 45, got %#v", policy["retention_days"])
	}

	listRec := app.get(t, "/api/v1/admin/retention-policies", adminCookies)
	requireStatus(t, listRec, http.StatusOK)
	policies := decodeSuccess[[]map[string]any](t, listRec)
	if len(policies) == 0 {
		t.Fatal("expected retention policy list to contain at least one policy")
	}
}

func TestAdmin_UserGetUpdateDeactivateFlow(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	adminCookies := app.login(t, admin)

	createRec := app.post(t, "/api/v1/admin/users", map[string]any{
		"email":        "user.mgmt@fitcommerce.test",
		"password":     defaultTestPassword,
		"role":         "coach",
		"display_name": "User Mgmt",
	}, adminCookies)
	requireStatus(t, createRec, http.StatusCreated)
	created := decodeSuccess[map[string]any](t, createRec)
	userID := created["id"].(string)

	getRec := app.get(t, "/api/v1/admin/users/"+userID, adminCookies)
	requireStatus(t, getRec, http.StatusOK)

	updateRec := app.put(t, "/api/v1/admin/users/"+userID, map[string]any{
		"display_name": "Updated User Mgmt",
		"role":         "operations_manager",
	}, adminCookies)
	requireStatus(t, updateRec, http.StatusOK)
	updated := decodeSuccess[map[string]any](t, updateRec)
	if updated["display_name"] != "Updated User Mgmt" {
		t.Fatalf("expected updated display_name, got %#v", updated["display_name"])
	}
	if updated["role"] != "operations_manager" {
		t.Fatalf("expected updated role operations_manager, got %#v", updated["role"])
	}

	deactivateRec := app.post(t, "/api/v1/admin/users/"+userID+"/deactivate", map[string]any{}, adminCookies)
	requireStatus(t, deactivateRec, http.StatusOK)

	getAfterRec := app.get(t, "/api/v1/admin/users/"+userID, adminCookies)
	requireStatus(t, getAfterRec, http.StatusOK)
	userAfter := decodeSuccess[map[string]any](t, getAfterRec)
	if userAfter["status"] != "inactive" {
		t.Fatalf("expected inactive status after deactivation, got %#v", userAfter["status"])
	}
}

func TestAdmin_UserEndpointValidationBranches(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	adminCookies := app.login(t, admin)

	createBadLocationRec := app.post(t, "/api/v1/admin/users", map[string]any{
		"email":        "bad.location@fitcommerce.test",
		"password":     defaultTestPassword,
		"role":         "coach",
		"display_name": "Bad Location",
		"location_id":  "not-a-uuid",
	}, adminCookies)
	requireStatus(t, createBadLocationRec, http.StatusBadRequest)

	getBadIDRec := app.get(t, "/api/v1/admin/users/not-a-uuid", adminCookies)
	requireStatus(t, getBadIDRec, http.StatusBadRequest)

	updateBadIDRec := app.put(t, "/api/v1/admin/users/not-a-uuid", map[string]any{
		"display_name": "x",
	}, adminCookies)
	requireStatus(t, updateBadIDRec, http.StatusBadRequest)

	deactivateBadIDRec := app.post(t, "/api/v1/admin/users/not-a-uuid/deactivate", map[string]any{}, adminCookies)
	requireStatus(t, deactivateBadIDRec, http.StatusBadRequest)

	createGoodUserRec := app.post(t, "/api/v1/admin/users", map[string]any{
		"email":        "valid.location@fitcommerce.test",
		"password":     defaultTestPassword,
		"role":         "coach",
		"display_name": "Valid Location",
	}, adminCookies)
	requireStatus(t, createGoodUserRec, http.StatusCreated)
	createdID := decodeSuccess[map[string]any](t, createGoodUserRec)["id"].(string)

	updateBadLocationRec := app.put(t, "/api/v1/admin/users/"+createdID, map[string]any{
		"location_id": "not-a-uuid",
	}, adminCookies)
	requireStatus(t, updateBadLocationRec, http.StatusBadRequest)
}

func TestAdmin_CreateUserWithLocationReturnsLocationID(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	adminCookies := app.login(t, admin)
	location := app.seedLocation(t, "North Branch")

	createRec := app.post(t, "/api/v1/admin/users", map[string]any{
		"email":        "with.location@fitcommerce.test",
		"password":     defaultTestPassword,
		"role":         "coach",
		"display_name": "With Location",
		"location_id":  location.ID.String(),
	}, adminCookies)
	requireStatus(t, createRec, http.StatusCreated)
	created := decodeSuccess[map[string]any](t, createRec)
	if created["location_id"] != location.ID.String() {
		t.Fatalf("expected location_id %s, got %#v", location.ID.String(), created["location_id"])
	}

	getRec := app.get(t, "/api/v1/admin/users/"+created["id"].(string), adminCookies)
	requireStatus(t, getRec, http.StatusOK)
	got := decodeSuccess[map[string]any](t, getRec)
	if got["location_id"] != location.ID.String() {
		t.Fatalf("expected persisted location_id %s, got %#v", location.ID.String(), got["location_id"])
	}
}
