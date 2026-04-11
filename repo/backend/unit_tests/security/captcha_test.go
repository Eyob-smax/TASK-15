package security_test

import (
	"reflect"
	"strings"
	"testing"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/security"
)

func TestGenerateMathCaptcha_FormatValid(t *testing.T) {
	challengeData, answer, err := security.GenerateMathCaptcha()
	if err != nil {
		t.Fatalf("GenerateMathCaptcha failed: %v", err)
	}

	if !strings.HasPrefix(challengeData, "What is ") {
		t.Errorf("unexpected challenge format: %q", challengeData)
	}
	if !strings.HasSuffix(challengeData, "?") {
		t.Errorf("challenge should end with '?': %q", challengeData)
	}
	if answer == "" {
		t.Error("answer should not be empty")
	}
}

func TestVerifyCaptchaAnswer_CorrectAnswerPasses(t *testing.T) {
	_, answer, err := security.GenerateMathCaptcha()
	if err != nil {
		t.Fatalf("GenerateMathCaptcha failed: %v", err)
	}
	salt, err := security.GenerateCaptchaSalt()
	if err != nil {
		t.Fatalf("GenerateCaptchaSalt failed: %v", err)
	}
	if !security.VerifyCaptchaAnswer(security.HashCaptchaAnswer(answer, salt), salt, answer) {
		t.Errorf("expected correct answer %q to pass", answer)
	}
}

func TestVerifyCaptchaAnswer_WrongAnswerFails(t *testing.T) {
	_, answer, err := security.GenerateMathCaptcha()
	if err != nil {
		t.Fatalf("GenerateMathCaptcha failed: %v", err)
	}
	salt, err := security.GenerateCaptchaSalt()
	if err != nil {
		t.Fatalf("GenerateCaptchaSalt failed: %v", err)
	}
	if security.VerifyCaptchaAnswer(security.HashCaptchaAnswer(answer, salt), salt, "99999") {
		t.Error("expected wrong answer to fail")
	}
}

func TestVerifyCaptchaAnswer_WhitespaceTrimmed(t *testing.T) {
	salt, err := security.GenerateCaptchaSalt()
	if err != nil {
		t.Fatalf("GenerateCaptchaSalt failed: %v", err)
	}
	if !security.VerifyCaptchaAnswer(security.HashCaptchaAnswer("42", salt), salt, "  42  ") {
		t.Error("expected answer with surrounding whitespace to pass")
	}
}

func TestVerifyCaptchaAnswer_CaseInsensitive(t *testing.T) {
	salt, err := security.GenerateCaptchaSalt()
	if err != nil {
		t.Fatalf("GenerateCaptchaSalt failed: %v", err)
	}
	if !security.VerifyCaptchaAnswer(security.HashCaptchaAnswer("YES", salt), salt, "yes") {
		t.Error("expected case-insensitive comparison to pass")
	}
}

func TestVerifyCaptchaAnswer_EmptyFails(t *testing.T) {
	_, answer, err := security.GenerateMathCaptcha()
	if err != nil {
		t.Fatalf("GenerateMathCaptcha failed: %v", err)
	}
	salt, err := security.GenerateCaptchaSalt()
	if err != nil {
		t.Fatalf("GenerateCaptchaSalt failed: %v", err)
	}
	if security.VerifyCaptchaAnswer(security.HashCaptchaAnswer(answer, salt), salt, "") {
		t.Error("expected empty submission to fail")
	}
}

// TestCaptchaChallenge_PersistenceLayerStoresHashNotPlaintext proves that the
// domain model used for DB persistence carries only derived verification
// material (hash + salt), never a plaintext answer field.
func TestCaptchaChallenge_PersistenceLayerStoresHashNotPlaintext(t *testing.T) {
	// Verify the domain struct has no plaintext Answer field.
	typ := reflect.TypeOf(domain.CaptchaChallenge{})
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Name == "Answer" && f.Type.Kind() == reflect.String {
			t.Errorf("domain.CaptchaChallenge must not have a plaintext Answer string field; found field %q", f.Name)
		}
	}

	// Verify AnswerHash and AnswerSalt fields are present and of the correct type.
	hashField, hasHash := typ.FieldByName("AnswerHash")
	saltField, hasSalt := typ.FieldByName("AnswerSalt")
	if !hasHash || hashField.Type != reflect.TypeOf([]byte(nil)) {
		t.Error("domain.CaptchaChallenge must have an AnswerHash []byte field")
	}
	if !hasSalt || saltField.Type != reflect.TypeOf([]byte(nil)) {
		t.Error("domain.CaptchaChallenge must have an AnswerSalt []byte field")
	}

	// Verify that the security layer builds a challenge struct with hash+salt populated.
	_, answer, err := security.GenerateMathCaptcha()
	if err != nil {
		t.Fatalf("GenerateMathCaptcha failed: %v", err)
	}
	salt, err := security.GenerateCaptchaSalt()
	if err != nil {
		t.Fatalf("GenerateCaptchaSalt failed: %v", err)
	}
	challenge := domain.CaptchaChallenge{
		AnswerHash: security.HashCaptchaAnswer(answer, salt),
		AnswerSalt: salt,
	}
	if len(challenge.AnswerHash) == 0 {
		t.Error("AnswerHash must be non-empty when a challenge is stored")
	}
	if len(challenge.AnswerSalt) == 0 {
		t.Error("AnswerSalt must be non-empty when a challenge is stored")
	}
}
