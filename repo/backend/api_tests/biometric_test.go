package api_tests

import (
	"net/http"
	"testing"

	"fitcommerce/internal/platform"
)

func biometricConfig(t *testing.T) platform.Config {
	t.Helper()
	cfg := integrationConfig(t)
	cfg.BiometricModuleEnabled = true
	cfg.BiometricMasterKeyRef = "integration-test-biometric-key"
	return cfg
}

func TestBiometric_ModuleDisabled_Returns501(t *testing.T) {
	app := newIntegrationApp(t) // BiometricModuleEnabled: false
	admin := app.seedUser(t, "administrator", nil)
	cookies := app.login(t, admin)
	targetUser := app.seedUser(t, "member", nil)

	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/admin/biometrics"},
		{http.MethodGet, "/api/v1/admin/biometrics/" + targetUser.ID.String()},
		{http.MethodPost, "/api/v1/admin/biometrics/" + targetUser.ID.String() + "/revoke"},
		{http.MethodPost, "/api/v1/admin/biometrics/rotate-key"},
		{http.MethodGet, "/api/v1/admin/biometrics/keys"},
	}

	for _, ep := range endpoints {
		var rec = app.post(t, ep.path, map[string]string{}, cookies)
		if ep.method == http.MethodGet {
			rec = app.get(t, ep.path, cookies)
		}
		if rec.Code != http.StatusNotImplemented {
			t.Errorf("%s %s: expected 501, got %d", ep.method, ep.path, rec.Code)
		}
	}
}

func TestBiometric_NonAdmin_Forbidden(t *testing.T) {
	app := newIntegrationAppWithConfig(t, biometricConfig(t))
	member := app.seedUser(t, "member", nil)
	cookies := app.login(t, member)
	targetUser := app.seedUser(t, "member", nil)

	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/admin/biometrics"},
		{http.MethodGet, "/api/v1/admin/biometrics/" + targetUser.ID.String()},
		{http.MethodPost, "/api/v1/admin/biometrics/" + targetUser.ID.String() + "/revoke"},
		{http.MethodPost, "/api/v1/admin/biometrics/rotate-key"},
		{http.MethodGet, "/api/v1/admin/biometrics/keys"},
	}

	for _, ep := range endpoints {
		if ep.method == http.MethodGet {
			rec := app.get(t, ep.path, cookies)
			if rec.Code != http.StatusForbidden {
				t.Errorf("%s %s: expected 403, got %d", ep.method, ep.path, rec.Code)
			}
		} else {
			rec := app.post(t, ep.path, map[string]string{}, cookies)
			if rec.Code != http.StatusForbidden {
				t.Errorf("%s %s: expected 403, got %d", ep.method, ep.path, rec.Code)
			}
		}
	}
}

func TestBiometric_Register_And_Get(t *testing.T) {
	app := newIntegrationAppWithConfig(t, biometricConfig(t))
	admin := app.seedUser(t, "administrator", nil)
	cookies := app.login(t, admin)
	targetUser := app.seedUser(t, "member", nil)

	// Register biometric for target user
	rec := app.post(t, "/api/v1/admin/biometrics", map[string]string{
		"user_id":      targetUser.ID.String(),
		"template_ref": "test-biometric-template-ref",
	}, cookies)
	requireStatus(t, rec, http.StatusCreated)

	body := decodeSuccess[map[string]any](t, rec)
	if body["user_id"] != targetUser.ID.String() {
		t.Fatalf("expected user_id %s, got %v", targetUser.ID.String(), body["user_id"])
	}
	if body["is_active"] != true {
		t.Fatalf("expected is_active true after register, got %v", body["is_active"])
	}

	// Get the enrollment — template_ref must be redacted
	getRec := app.get(t, "/api/v1/admin/biometrics/"+targetUser.ID.String(), cookies)
	requireStatus(t, getRec, http.StatusOK)

	getBody := decodeSuccess[map[string]any](t, getRec)
	if getBody["template_ref"] == "test-biometric-template-ref" {
		t.Fatal("expected template_ref to be redacted, but got plaintext value")
	}
	if getBody["is_active"] != true {
		t.Fatalf("expected is_active true on GET, got %v", getBody["is_active"])
	}
}

func TestBiometric_Revoke(t *testing.T) {
	app := newIntegrationAppWithConfig(t, biometricConfig(t))
	admin := app.seedUser(t, "administrator", nil)
	cookies := app.login(t, admin)
	targetUser := app.seedUser(t, "member", nil)

	// Register first
	rec := app.post(t, "/api/v1/admin/biometrics", map[string]string{
		"user_id":      targetUser.ID.String(),
		"template_ref": "revoke-test-ref",
	}, cookies)
	requireStatus(t, rec, http.StatusCreated)

	// Revoke
	revokeRec := app.post(t, "/api/v1/admin/biometrics/"+targetUser.ID.String()+"/revoke", map[string]string{}, cookies)
	requireStatus(t, revokeRec, http.StatusOK)

	// Get after revoke — is_active should be false
	getRec := app.get(t, "/api/v1/admin/biometrics/"+targetUser.ID.String(), cookies)
	requireStatus(t, getRec, http.StatusOK)

	getBody := decodeSuccess[map[string]any](t, getRec)
	if getBody["is_active"] != false {
		t.Fatalf("expected is_active false after revoke, got %v", getBody["is_active"])
	}
}

func TestBiometric_ListKeys(t *testing.T) {
	app := newIntegrationAppWithConfig(t, biometricConfig(t))
	admin := app.seedUser(t, "administrator", nil)
	cookies := app.login(t, admin)
	targetUser := app.seedUser(t, "member", nil)

	// Register to ensure an active key exists
	rec := app.post(t, "/api/v1/admin/biometrics", map[string]string{
		"user_id":      targetUser.ID.String(),
		"template_ref": "listkeys-test-ref",
	}, cookies)
	requireStatus(t, rec, http.StatusCreated)

	// List keys
	keysRec := app.get(t, "/api/v1/admin/biometrics/keys", cookies)
	requireStatus(t, keysRec, http.StatusOK)

	keys := decodeSuccess[[]map[string]any](t, keysRec)
	if len(keys) == 0 {
		t.Fatal("expected at least one encryption key after register, got none")
	}
}

func TestBiometric_RotateKey(t *testing.T) {
	app := newIntegrationAppWithConfig(t, biometricConfig(t))
	admin := app.seedUser(t, "administrator", nil)
	cookies := app.login(t, admin)

	// Rotate key
	rec := app.post(t, "/api/v1/admin/biometrics/rotate-key", map[string]string{}, cookies)
	requireStatus(t, rec, http.StatusOK)

	body := decodeSuccess[map[string]any](t, rec)
	if body["id"] == nil || body["id"] == "" {
		t.Fatalf("expected new key id in rotate-key response, got %v", body)
	}
	if body["is_active"] != true {
		t.Fatalf("expected rotated key to be active, got is_active=%v", body["is_active"])
	}
}
