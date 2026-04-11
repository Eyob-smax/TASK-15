package api_tests

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// triggerCaptchaChallenge is shared setup: locks the user out via 5 bad
// logins, immediately expires the lockout window via direct SQL, then
// attempts a login to receive a CAPTCHA_REQUIRED response.
// Returns the challenge_id string from the error detail.
func triggerCaptchaChallenge(t *testing.T, app *integrationApp) string {
	t.Helper()

	user := app.seedUser(t, "administrator", nil)

	for i := 0; i < 5; i++ {
		rec := app.post(t, "/api/v1/auth/login", map[string]string{
			"email":    user.Email,
			"password": "WrongPassword123!",
		}, nil)
		requireStatus(t, rec, http.StatusUnauthorized)
	}

	expiredLock := time.Now().UTC().Add(-1 * time.Minute)
	if _, err := app.app.Pool.Exec(context.Background(),
		`UPDATE users SET locked_until = $2 WHERE id = $1`,
		user.ID, expiredLock,
	); err != nil {
		t.Fatalf("expire lockout: %v", err)
	}

	captchaRec := app.post(t, "/api/v1/auth/login", map[string]string{
		"email":    user.Email,
		"password": user.Password,
	}, nil)
	requireStatus(t, captchaRec, http.StatusForbidden)

	errBody := decodeError(t, captchaRec)
	if errBody.Error.Code != "CAPTCHA_REQUIRED" {
		t.Fatalf("expected CAPTCHA_REQUIRED, got %s", errBody.Error.Code)
	}

	var challengeID string
	for _, detail := range errBody.Error.Details {
		if detail.Field == "challenge_id" {
			challengeID = detail.Message
		}
	}
	if challengeID == "" {
		t.Fatal("expected challenge_id in captcha-required error details")
	}
	return challengeID
}

func TestAuth_LockoutAfterFiveFailures(t *testing.T) {
	app := newIntegrationApp(t)
	user := app.seedUser(t, "administrator", nil)

	for i := 0; i < 5; i++ {
		rec := app.post(t, "/api/v1/auth/login", map[string]string{
			"email":    user.Email,
			"password": "WrongPassword123!",
		}, nil)
		requireStatus(t, rec, http.StatusUnauthorized)
	}

	lockedRec := app.post(t, "/api/v1/auth/login", map[string]string{
		"email":    user.Email,
		"password": user.Password,
	}, nil)
	requireStatus(t, lockedRec, http.StatusLocked)

	errBody := decodeError(t, lockedRec)
	if errBody.Error.Code != "ACCOUNT_LOCKED" {
		t.Fatalf("expected ACCOUNT_LOCKED, got %s", errBody.Error.Code)
	}
}

func TestAuth_CaptchaFlowAfterLockoutExpiry(t *testing.T) {
	app := newIntegrationApp(t)
	challengeID := triggerCaptchaChallenge(t, app)

	wrongAnswerRec := app.post(t, "/api/v1/auth/captcha/verify", map[string]string{
		"challenge_id": challengeID,
		"answer":       "99999",
	}, nil)
	requireStatus(t, wrongAnswerRec, http.StatusUnprocessableEntity)

	verifyRec := app.post(t, "/api/v1/auth/captcha/verify", map[string]string{
		"challenge_id": challengeID,
		"answer":       app.captchaAnswer(t, challengeID),
	}, nil)
	requireStatus(t, verifyRec, http.StatusOK)

	missingChallengeRec := app.post(t, "/api/v1/auth/captcha/verify", map[string]string{
		"challenge_id": "11111111-1111-1111-1111-111111111111",
		"answer":       "42",
	}, nil)
	requireStatus(t, missingChallengeRec, http.StatusUnauthorized)
}

// TestAuth_CaptchaChallenge_DBRowHasHashNotPlaintext is the postgres-boundary
// persistence assertion for G1: it proves that the captcha_challenges table
// stores derived verification material (answer_hash + answer_salt) and that
// no plaintext answer column exists at the DB level.
func TestAuth_CaptchaChallenge_DBRowHasHashNotPlaintext(t *testing.T) {
	app := newIntegrationApp(t)
	challengeID := triggerCaptchaChallenge(t, app)

	// Assert the DB row carries hash + salt, both non-empty.
	var answerHash, answerSalt []byte
	err := app.app.Pool.QueryRow(context.Background(),
		`SELECT answer_hash, answer_salt FROM captcha_challenges WHERE id = $1`,
		challengeID,
	).Scan(&answerHash, &answerSalt)
	if err != nil {
		t.Fatalf("query captcha_challenges row: %v", err)
	}
	if len(answerHash) == 0 {
		t.Error("answer_hash column must be non-empty in persisted DB row")
	}
	if len(answerSalt) == 0 {
		t.Error("answer_salt column must be non-empty in persisted DB row")
	}

	// Assert no plaintext answer column exists in the table schema.
	var plaintextColumnExists bool
	err = app.app.Pool.QueryRow(context.Background(),
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name   = 'captcha_challenges'
			  AND column_name  = 'answer'
		)`,
	).Scan(&plaintextColumnExists)
	if err != nil {
		t.Fatalf("check information_schema for plaintext answer column: %v", err)
	}
	if plaintextColumnExists {
		t.Error("captcha_challenges table must not contain a plaintext 'answer' column")
	}
}
