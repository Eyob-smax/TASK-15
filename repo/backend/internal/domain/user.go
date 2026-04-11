package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents an authenticated user of the system.
type User struct {
	ID               uuid.UUID
	Email            string
	PasswordHash     string
	Salt             string
	Role             UserRole
	Status           UserStatus
	DisplayName      string
	LocationID       *uuid.UUID
	FailedLoginCount int
	LockedUntil      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// IsLocked returns true if the user account is currently locked
// either by status or by a time-based lockout that has not yet expired.
func (u *User) IsLocked() bool {
	if u.Status == UserStatusLocked {
		return true
	}
	if u.LockedUntil != nil && time.Now().Before(*u.LockedUntil) {
		return true
	}
	return false
}

// IncrementFailedLogin increments the failed login counter and updates the timestamp.
func (u *User) IncrementFailedLogin() {
	u.FailedLoginCount++
	now := time.Now()
	u.UpdatedAt = now
}

// ResetFailedLogin resets the failed login counter to zero.
func (u *User) ResetFailedLogin() {
	u.FailedLoginCount = 0
	u.LockedUntil = nil
	if u.Status == UserStatusLocked {
		u.Status = UserStatusActive
	}
	u.UpdatedAt = time.Now()
}

// Lock locks the user account for the specified duration.
func (u *User) Lock(duration time.Duration) {
	u.Status = UserStatusLocked
	lockUntil := time.Now().Add(duration)
	u.LockedUntil = &lockUntil
	u.UpdatedAt = time.Now()
}

// Session represents an active user session.
type Session struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	Token             string
	IdleExpiresAt     time.Time
	AbsoluteExpiresAt time.Time
	CreatedAt         time.Time
}

// IsExpired returns true if the session has expired either by idle timeout
// or by absolute timeout.
func (s *Session) IsExpired() bool {
	now := time.Now()
	return now.After(s.IdleExpiresAt) || now.After(s.AbsoluteExpiresAt)
}

// RefreshIdle extends the idle expiration by the given timeout duration,
// but never beyond the absolute expiration.
func (s *Session) RefreshIdle(timeout time.Duration) {
	newIdle := time.Now().Add(timeout)
	if newIdle.After(s.AbsoluteExpiresAt) {
		s.IdleExpiresAt = s.AbsoluteExpiresAt
	} else {
		s.IdleExpiresAt = newIdle
	}
}

// CaptchaChallenge represents a CAPTCHA challenge presented to a user
// after repeated failed login attempts.
type CaptchaChallenge struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	ChallengeData string
	AnswerHash    []byte
	AnswerSalt    []byte
	CreatedAt     time.Time
	ExpiresAt     time.Time
	Verified      bool
}
