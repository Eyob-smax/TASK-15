package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/platform"
	"fitcommerce/internal/security"
	"fitcommerce/internal/store"
)

// AuthServiceImpl is the concrete implementation of AuthService. It owns the
// login lockout flow, session lifecycle, and CAPTCHA gating.
type AuthServiceImpl struct {
	users    store.UserRepository
	sessions store.SessionRepository
	captchas store.CaptchaRepository
	audit    AuditService
	cfg      *platform.Config
}

// NewAuthService constructs an AuthServiceImpl with all required dependencies.
func NewAuthService(
	users store.UserRepository,
	sessions store.SessionRepository,
	captchas store.CaptchaRepository,
	audit AuditService,
	cfg *platform.Config,
) *AuthServiceImpl {
	return &AuthServiceImpl{
		users:    users,
		sessions: sessions,
		captchas: captchas,
		audit:    audit,
		cfg:      cfg,
	}
}

// Login authenticates a user by email and password, enforcing the lockout and
// CAPTCHA gate. Returns the new session and authenticated user on success.
//
// Flow:
//  1. Look up the user — not found returns ErrUnauthorized (no user-existence leak).
//  2. If locked_until is in the future → ErrAccountLocked.
//  3. If locked_until is in the past (lockout window just expired) → generate CAPTCHA
//     → ErrCaptchaRequired (client must call VerifyCaptcha before retrying).
//  4. Verify password:
//     - Wrong → increment counter; if ≥ threshold → Lock → audit LoginLockout → ErrUnauthorized.
//     - Correct → reset counter → create session → audit LoginSuccess → return.
func (s *AuthServiceImpl) Login(ctx context.Context, email, password string) (*domain.Session, *domain.User, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		// Do not reveal whether the email exists.
		platform.LogLoginAttempt(slog.Default(), email, false, "invalid_credentials")
		return nil, nil, domain.ErrUnauthorized
	}

	// Inactive accounts (including the reserved system actor) can never log in.
	if user.Status == domain.UserStatusInactive {
		platform.LogLoginAttempt(slog.Default(), email, false, "inactive_account")
		return nil, nil, domain.ErrUnauthorized
	}

	now := time.Now()

	if user.LockedUntil != nil {
		if now.Before(*user.LockedUntil) {
			// Active lockout — account is still locked.
			platform.LogLoginAttempt(slog.Default(), email, false, "account_locked")
			platform.LogLockout(slog.Default(), email, *user.LockedUntil)
			return nil, nil, &domain.ErrAccountLocked{
				Message:     "account is temporarily locked",
				LockedUntil: user.LockedUntil,
			}
		}
		// Lockout window has elapsed — require CAPTCHA before allowing a new attempt.
		challengeData, answer, err := security.GenerateMathCaptcha()
		if err != nil {
			return nil, nil, fmt.Errorf("auth_service.Login generate captcha: %w", err)
		}
		answerSalt, err := security.GenerateCaptchaSalt()
		if err != nil {
			return nil, nil, fmt.Errorf("auth_service.Login generate captcha salt: %w", err)
		}
		challenge := &domain.CaptchaChallenge{
			ID:            uuid.New(),
			UserID:        user.ID,
			ChallengeData: challengeData,
			AnswerHash:    security.HashCaptchaAnswer(answer, answerSalt),
			AnswerSalt:    answerSalt,
			CreatedAt:     now,
			ExpiresAt:     now.Add(10 * time.Minute),
			Verified:      false,
		}
		if err := s.captchas.Create(ctx, challenge); err != nil {
			return nil, nil, fmt.Errorf("auth_service.Login create captcha: %w", err)
		}
		platform.LogLoginAttempt(slog.Default(), email, false, "captcha_required")
		return nil, nil, &domain.ErrCaptchaRequired{
			ChallengeID:   challenge.ID.String(),
			ChallengeData: challenge.ChallengeData,
		}
	}

	// No active lockout — verify the password.
	match, err := security.VerifyPassword(password, user.PasswordHash, user.Salt)
	if err != nil {
		return nil, nil, fmt.Errorf("auth_service.Login verify password: %w", err)
	}

	if !match {
		user.IncrementFailedLogin()
		threshold := s.cfg.LoginLockoutThreshold
		var auditErr error
		if user.FailedLoginCount >= threshold {
			lockDuration := time.Duration(s.cfg.LoginLockoutDurationMinutes) * time.Minute
			user.Lock(lockDuration)
			if err := s.users.Update(ctx, user); err != nil {
				return nil, nil, fmt.Errorf("auth_service.Login lock user: %w", err)
			}
			auditErr = s.audit.Log(ctx, security.EventLoginLockout, "user", user.ID, user.ID,
				map[string]interface{}{"email": user.Email, "reason": "failed_login_threshold"})
			if user.LockedUntil != nil {
				platform.LogLockout(slog.Default(), email, *user.LockedUntil)
			}
		} else {
			if err := s.users.Update(ctx, user); err != nil {
				return nil, nil, fmt.Errorf("auth_service.Login update failed count: %w", err)
			}
			auditErr = s.audit.Log(ctx, security.EventLoginFailure, "user", user.ID, user.ID,
				map[string]interface{}{"email": user.Email, "failed_count": user.FailedLoginCount})
		}
		_ = auditErr // audit failures are non-fatal; login response takes priority
		platform.LogLoginAttempt(slog.Default(), email, false, "invalid_credentials")
		return nil, nil, domain.ErrUnauthorized
	}

	// Password correct — create session.
	user.ResetFailedLogin()
	if err := s.users.Update(ctx, user); err != nil {
		return nil, nil, fmt.Errorf("auth_service.Login reset failed login: %w", err)
	}

	token, err := security.GenerateSessionToken()
	if err != nil {
		return nil, nil, fmt.Errorf("auth_service.Login generate token: %w", err)
	}

	idleTimeout := time.Duration(s.cfg.SessionIdleTimeoutMinutes) * time.Minute
	absoluteTimeout := time.Duration(s.cfg.SessionAbsoluteTimeoutHours) * time.Hour

	session := &domain.Session{
		ID:                uuid.New(),
		UserID:            user.ID,
		Token:             token,
		IdleExpiresAt:     now.Add(idleTimeout),
		AbsoluteExpiresAt: now.Add(absoluteTimeout),
		CreatedAt:         now,
	}
	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, nil, fmt.Errorf("auth_service.Login create session: %w", err)
	}

	_ = s.audit.Log(ctx, security.EventLoginSuccess, "user", user.ID, user.ID,
		map[string]interface{}{"email": user.Email})
	platform.LogLoginAttempt(slog.Default(), email, true, "")

	return session, user, nil
}

// Logout invalidates the session identified by token and emits an audit event.
func (s *AuthServiceImpl) Logout(ctx context.Context, token string) error {
	sess, err := s.sessions.GetByToken(ctx, token)
	if err != nil {
		// Already gone — treat as success.
		return nil
	}

	if err := s.sessions.Delete(ctx, sess.ID); err != nil {
		return fmt.Errorf("auth_service.Logout delete session: %w", err)
	}

	_ = s.audit.Log(ctx, security.EventLogout, "session", sess.ID, sess.UserID,
		map[string]interface{}{})

	return nil
}

// ValidateSession verifies that a session token is current, refreshes the idle
// timeout, and returns the session and its owning user. Returns ErrUnauthorized
// if the token is missing, expired (absolute or idle), or the user cannot be found.
func (s *AuthServiceImpl) ValidateSession(ctx context.Context, token string) (*domain.Session, *domain.User, error) {
	sess, err := s.sessions.GetByToken(ctx, token)
	if err != nil {
		return nil, nil, domain.ErrUnauthorized
	}

	now := time.Now()

	if now.After(sess.AbsoluteExpiresAt) {
		_ = s.sessions.Delete(ctx, sess.ID)
		_ = s.audit.Log(ctx, security.EventSessionExpired, "session", sess.ID, sess.UserID,
			map[string]interface{}{"reason": "absolute_timeout"})
		platform.LogSessionExpired(slog.Default(), sess.UserID, "absolute_timeout")
		return nil, nil, domain.ErrUnauthorized
	}

	if now.After(sess.IdleExpiresAt) {
		_ = s.sessions.Delete(ctx, sess.ID)
		_ = s.audit.Log(ctx, security.EventSessionExpired, "session", sess.ID, sess.UserID,
			map[string]interface{}{"reason": "idle_timeout"})
		platform.LogSessionExpired(slog.Default(), sess.UserID, "idle_timeout")
		return nil, nil, domain.ErrUnauthorized
	}

	// Refresh idle expiry.
	idleTimeout := time.Duration(s.cfg.SessionIdleTimeoutMinutes) * time.Minute
	sess.RefreshIdle(idleTimeout)
	if err := s.sessions.UpdateIdleExpiry(ctx, sess.ID, sess.IdleExpiresAt); err != nil {
		return nil, nil, fmt.Errorf("auth_service.ValidateSession update idle expiry: %w", err)
	}

	user, err := s.users.GetByID(ctx, sess.UserID)
	if err != nil {
		return nil, nil, domain.ErrUnauthorized
	}

	return sess, user, nil
}

// VerifyCaptcha validates a submitted CAPTCHA answer. On success it marks the
// challenge as verified and resets the user's lockout state so they can retry
// login normally.
func (s *AuthServiceImpl) VerifyCaptcha(ctx context.Context, challengeID uuid.UUID, answer string) error {
	challenge, err := s.captchas.GetByID(ctx, challengeID)
	if err != nil {
		return domain.ErrUnauthorized
	}

	now := time.Now()
	if now.After(challenge.ExpiresAt) {
		return domain.ErrUnauthorized
	}
	if challenge.Verified {
		return domain.ErrUnauthorized
	}

	if !security.VerifyCaptchaAnswer(challenge.AnswerHash, challenge.AnswerSalt, answer) {
		_ = s.audit.Log(ctx, security.EventCaptchaFailed, "captcha_challenge", challenge.ID, challenge.UserID,
			map[string]interface{}{})
		return &domain.ErrValidation{Field: "answer", Message: "incorrect captcha answer"}
	}

	if err := s.captchas.MarkVerified(ctx, challengeID); err != nil {
		return fmt.Errorf("auth_service.VerifyCaptcha mark verified: %w", err)
	}

	// Reset user lockout so they can attempt login again.
	user, err := s.users.GetByID(ctx, challenge.UserID)
	if err == nil {
		user.ResetFailedLogin()
		_ = s.users.Update(ctx, user)
	}

	_ = s.audit.Log(ctx, security.EventCaptchaVerified, "captcha_challenge", challenge.ID, challenge.UserID,
		map[string]interface{}{})

	return nil
}

// Compile-time interface assertion.
var _ AuthService = (*AuthServiceImpl)(nil)
