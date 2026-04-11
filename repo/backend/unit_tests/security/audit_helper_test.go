package security_test

import (
	"testing"

	"github.com/google/uuid"

	"fitcommerce/internal/security"
)

func TestRedactSensitiveFields_RemovesRestrictedKeys(t *testing.T) {
	input := map[string]interface{}{
		"email":          "alice@example.com",
		"password":       "secret",
		"password_hash":  "hash",
		"salt":           "nacl",
		"token":          "tok",
		"session_token":  "sess",
		"answer":         "42",
		"key":            "k",
		"secret":         "s",
		"encrypted_data": "enc",
		"event_type":     "auth.login.success",
	}

	out := security.RedactSensitiveFields(input)

	restricted := []string{"password", "password_hash", "salt", "token", "session_token", "answer", "key", "secret", "encrypted_data"}
	for _, k := range restricted {
		if _, exists := out[k]; exists {
			t.Errorf("RedactSensitiveFields: key %q should have been removed", k)
		}
	}

	allowed := []string{"email", "event_type"}
	for _, k := range allowed {
		if _, exists := out[k]; !exists {
			t.Errorf("RedactSensitiveFields: key %q should be retained", k)
		}
	}
}

func TestSafeDetails_IsIdempotent(t *testing.T) {
	details := map[string]interface{}{
		"email":    "alice@example.com",
		"password": "should_be_removed",
	}
	once := security.SafeDetails(details)
	twice := security.SafeDetails(once)

	if _, exists := twice["password"]; exists {
		t.Error("SafeDetails is not idempotent: password survived two passes")
	}
	if twice["email"] != "alice@example.com" {
		t.Error("SafeDetails should retain non-sensitive keys")
	}
}

func TestBuildAuditEvent_HashChain(t *testing.T) {
	actorID := uuid.New()
	entityID := uuid.New()

	event1 := security.BuildAuditEvent(
		"auth.login.success", "user",
		entityID, actorID,
		map[string]interface{}{"email": "alice@example.com"},
		"", // no predecessor
	)
	if event1.IntegrityHash == "" {
		t.Fatal("event1 IntegrityHash should not be empty")
	}
	if event1.PreviousHash != "" {
		t.Errorf("event1 PreviousHash should be empty, got %q", event1.PreviousHash)
	}

	event2 := security.BuildAuditEvent(
		"auth.logout", "session",
		entityID, actorID,
		map[string]interface{}{},
		event1.IntegrityHash,
	)
	if event2.PreviousHash != event1.IntegrityHash {
		t.Errorf("event2.PreviousHash = %q, want %q", event2.PreviousHash, event1.IntegrityHash)
	}
	if event2.IntegrityHash == event1.IntegrityHash {
		t.Error("consecutive events should have different integrity hashes")
	}
}

func TestBuildAuditEvent_ConstantsExist(t *testing.T) {
	constants := []string{
		security.EventLoginSuccess,
		security.EventLoginFailure,
		security.EventLoginLockout,
		security.EventLogout,
		security.EventSessionExpired,
		security.EventCaptchaVerified,
		security.EventCaptchaFailed,
	}
	for _, c := range constants {
		if c == "" {
			t.Error("audit event constant must not be empty")
		}
	}
}
