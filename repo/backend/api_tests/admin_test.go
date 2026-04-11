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
}
