package security_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/security"
)

func TestMaskEmail_Basic(t *testing.T) {
	got := security.MaskEmail("alice@example.com")
	if !strings.HasPrefix(got, "a***@") {
		t.Errorf("MaskEmail: expected prefix 'a***@', got %q", got)
	}
	if !strings.HasSuffix(got, "example.com") {
		t.Errorf("MaskEmail: expected suffix 'example.com', got %q", got)
	}
}

func TestMaskEmail_ShortLocal(t *testing.T) {
	got := security.MaskEmail("a@b.com")
	if !strings.HasPrefix(got, "a***@") {
		t.Errorf("MaskEmail short local: got %q", got)
	}
}

func TestMaskEmail_NoAt(t *testing.T) {
	got := security.MaskEmail("notanemail")
	if got == "" {
		t.Error("MaskEmail with no @ should return a placeholder, not empty")
	}
}

func TestMaskPhone_LastFourVisible(t *testing.T) {
	got := security.MaskPhone("+1-555-867-5309")
	if !strings.HasSuffix(got, "5309") {
		t.Errorf("MaskPhone: expected suffix '5309', got %q", got)
	}
	if strings.Contains(got, "5") && !strings.HasSuffix(got, "5309") {
		t.Errorf("MaskPhone leaked non-last-4 digits: %q", got)
	}
}

func TestMaskPhone_AllStars(t *testing.T) {
	got := security.MaskPhone("+1-555-867-5309")
	// Everything before the last 4 digits should be *
	prefix := got[:len(got)-4]
	for _, ch := range prefix {
		if ch != '*' {
			t.Errorf("MaskPhone prefix should be all '*', got %q", prefix)
			break
		}
	}
}

func TestRedactBiometric(t *testing.T) {
	got := security.RedactBiometric()
	if got != "[BIOMETRIC REDACTED]" {
		t.Errorf("unexpected redaction placeholder: %q", got)
	}
}

func TestMaskFieldByRole_AdminGetsFullEmail(t *testing.T) {
	id := uuid.New()
	got := security.MaskFieldByRole("email", "admin@example.com", domain.UserRoleAdministrator, id, id)
	if got != "admin@example.com" {
		t.Errorf("admin should see full email, got %q", got)
	}
}

func TestMaskFieldByRole_MemberOwnEmail(t *testing.T) {
	id := uuid.New()
	got := security.MaskFieldByRole("email", "member@example.com", domain.UserRoleMember, id, id)
	if got != "member@example.com" {
		t.Errorf("member viewing own email should see full value, got %q", got)
	}
}

func TestMaskFieldByRole_MemberOtherEmail(t *testing.T) {
	ownerID := uuid.New()
	requesterID := uuid.New()
	got := security.MaskFieldByRole("email", "other@example.com", domain.UserRoleMember, ownerID, requesterID)
	if got == "other@example.com" {
		t.Error("member viewing other user's email should get masked value")
	}
}

func TestMaskFieldByRole_PasswordHashNeverReturned(t *testing.T) {
	id := uuid.New()
	for _, role := range domain.AllUserRoles() {
		got := security.MaskFieldByRole("password_hash", "secret", role, id, id)
		if got != "" {
			t.Errorf("password_hash should always return empty for role %s, got %q", role, got)
		}
	}
}

func TestMaskFieldByRole_BiometricAlwaysRedacted(t *testing.T) {
	id := uuid.New()
	for _, role := range domain.AllUserRoles() {
		got := security.MaskFieldByRole("biometric", "rawdata", role, id, id)
		if got != security.RedactBiometric() {
			t.Errorf("biometric should always be redacted for role %s", role)
		}
	}
}
