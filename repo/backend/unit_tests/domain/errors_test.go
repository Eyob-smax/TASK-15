package domain_test

import (
	"strings"
	"testing"
	"time"

	"fitcommerce/internal/domain"
)

func TestErrConflict_Error(t *testing.T) {
	withMsg := (&domain.ErrConflict{Entity: "user", Message: "dup email"}).Error()
	if !strings.Contains(withMsg, "conflict on user") || !strings.Contains(withMsg, "dup email") {
		t.Errorf("unexpected: %q", withMsg)
	}
	noMsg := (&domain.ErrConflict{Entity: "user"}).Error()
	if !strings.Contains(noMsg, "conflict on user") {
		t.Errorf("unexpected: %q", noMsg)
	}
	if strings.Contains(noMsg, ":") {
		t.Errorf("should not include colon suffix: %q", noMsg)
	}
}

func TestErrValidation_Error(t *testing.T) {
	withField := (&domain.ErrValidation{Field: "email", Message: "bad"}).Error()
	if !strings.Contains(withField, "email") || !strings.Contains(withField, "bad") {
		t.Errorf("unexpected: %q", withField)
	}
	noField := (&domain.ErrValidation{Message: "bad"}).Error()
	if !strings.Contains(noField, "validation error") || !strings.Contains(noField, "bad") {
		t.Errorf("unexpected: %q", noField)
	}
}

func TestValidationError_Error(t *testing.T) {
	got := domain.ValidationError{Field: "name", Message: "required"}.Error()
	if got != "name: required" {
		t.Errorf("got %q", got)
	}
}

func TestErrAccountLocked_Error(t *testing.T) {
	until := time.Now().Add(5 * time.Minute)
	withMsg := (&domain.ErrAccountLocked{Message: "too many attempts", LockedUntil: &until}).Error()
	if !strings.Contains(withMsg, "account locked") || !strings.Contains(withMsg, "too many attempts") {
		t.Errorf("unexpected: %q", withMsg)
	}
	noMsg := (&domain.ErrAccountLocked{}).Error()
	if !strings.Contains(noMsg, "account locked") {
		t.Errorf("unexpected: %q", noMsg)
	}
}

func TestErrCaptchaRequired_Error(t *testing.T) {
	got := (&domain.ErrCaptchaRequired{ChallengeID: "abc123"}).Error()
	if !strings.Contains(got, "captcha required") || !strings.Contains(got, "abc123") {
		t.Errorf("unexpected: %q", got)
	}
}

func TestErrInvalidTransition_Error(t *testing.T) {
	got := (&domain.ErrInvalidTransition{Entity: "order", From: "draft", To: "shipped"}).Error()
	for _, s := range []string{"order", "draft", "shipped"} {
		if !strings.Contains(got, s) {
			t.Errorf("expected %q in %q", s, got)
		}
	}
}

func TestErrPublishBlocked_Error(t *testing.T) {
	got := (&domain.ErrPublishBlocked{Reasons: []string{"no title", "no price"}}).Error()
	for _, s := range []string{"publish blocked", "no title", "no price"} {
		if !strings.Contains(got, s) {
			t.Errorf("expected %q in %q", s, got)
		}
	}
}

func TestErrBatchEditPartialFailure_Error(t *testing.T) {
	got := (&domain.ErrBatchEditPartialFailure{FailedRows: []domain.BatchEditFailedRow{{}, {}}}).Error()
	if !strings.Contains(got, "2 rows failed") {
		t.Errorf("expected count in message, got %q", got)
	}
}

func TestErrVarianceUnresolved_Error(t *testing.T) {
	got := (&domain.ErrVarianceUnresolved{VarianceID: "v-42"}).Error()
	if !strings.Contains(got, "v-42") {
		t.Errorf("expected ID in %q", got)
	}
}

func TestErrRetentionViolation_Error(t *testing.T) {
	got := (&domain.ErrRetentionViolation{EntityType: "order", RetentionDays: 90}).Error()
	for _, s := range []string{"order", "90"} {
		if !strings.Contains(got, s) {
			t.Errorf("expected %q in %q", s, got)
		}
	}
}
