package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
)

func TestUser_IsLocked_ByStatus(t *testing.T) {
	u := &domain.User{Status: domain.UserStatusLocked}
	if !u.IsLocked() {
		t.Error("expected locked when status is locked")
	}
}

func TestUser_IsLocked_ByLockoutTime_Future(t *testing.T) {
	future := time.Now().Add(10 * time.Minute)
	u := &domain.User{Status: domain.UserStatusActive, LockedUntil: &future}
	if !u.IsLocked() {
		t.Error("expected locked when LockedUntil is in the future")
	}
}

func TestUser_IsLocked_ByLockoutTime_Past(t *testing.T) {
	past := time.Now().Add(-10 * time.Minute)
	u := &domain.User{Status: domain.UserStatusActive, LockedUntil: &past}
	if u.IsLocked() {
		t.Error("expected unlocked when LockedUntil is in the past")
	}
}

func TestUser_IsLocked_NilLockoutAndActive(t *testing.T) {
	u := &domain.User{Status: domain.UserStatusActive}
	if u.IsLocked() {
		t.Error("expected not locked when status active and LockedUntil nil")
	}
}

func TestUser_IncrementFailedLogin(t *testing.T) {
	u := &domain.User{FailedLoginCount: 2}
	before := u.UpdatedAt
	u.IncrementFailedLogin()
	if u.FailedLoginCount != 3 {
		t.Errorf("expected 3, got %d", u.FailedLoginCount)
	}
	if !u.UpdatedAt.After(before) {
		t.Error("expected UpdatedAt to advance")
	}
}

func TestUser_ResetFailedLogin_ClearsLock(t *testing.T) {
	lockedAt := time.Now().Add(time.Minute)
	u := &domain.User{
		FailedLoginCount: 5,
		LockedUntil:      &lockedAt,
		Status:           domain.UserStatusLocked,
	}
	u.ResetFailedLogin()
	if u.FailedLoginCount != 0 {
		t.Errorf("expected 0 fails, got %d", u.FailedLoginCount)
	}
	if u.LockedUntil != nil {
		t.Error("expected LockedUntil to be nil")
	}
	if u.Status != domain.UserStatusActive {
		t.Errorf("expected status reset to active, got %s", u.Status)
	}
}

func TestUser_ResetFailedLogin_KeepsNonLockedStatus(t *testing.T) {
	u := &domain.User{FailedLoginCount: 1, Status: domain.UserStatusInactive}
	u.ResetFailedLogin()
	if u.Status != domain.UserStatusInactive {
		t.Errorf("non-locked status should be preserved, got %s", u.Status)
	}
}

func TestUser_Lock(t *testing.T) {
	u := &domain.User{Status: domain.UserStatusActive}
	u.Lock(30 * time.Minute)
	if u.Status != domain.UserStatusLocked {
		t.Errorf("expected locked, got %s", u.Status)
	}
	if u.LockedUntil == nil {
		t.Fatal("expected LockedUntil set")
	}
	if !u.LockedUntil.After(time.Now()) {
		t.Error("expected LockedUntil in the future")
	}
}

func TestSession_IsExpired_IdleTimeout(t *testing.T) {
	s := &domain.Session{
		ID:                uuid.New(),
		IdleExpiresAt:     time.Now().Add(-time.Second),
		AbsoluteExpiresAt: time.Now().Add(time.Hour),
	}
	if !s.IsExpired() {
		t.Error("expected expired when idle is past")
	}
}

func TestSession_IsExpired_AbsoluteTimeout(t *testing.T) {
	s := &domain.Session{
		IdleExpiresAt:     time.Now().Add(time.Hour),
		AbsoluteExpiresAt: time.Now().Add(-time.Second),
	}
	if !s.IsExpired() {
		t.Error("expected expired when absolute is past")
	}
}

func TestSession_IsExpired_Active(t *testing.T) {
	s := &domain.Session{
		IdleExpiresAt:     time.Now().Add(time.Minute),
		AbsoluteExpiresAt: time.Now().Add(time.Hour),
	}
	if s.IsExpired() {
		t.Error("expected not expired")
	}
}

func TestSession_RefreshIdle_WithinAbsolute(t *testing.T) {
	abs := time.Now().Add(time.Hour)
	s := &domain.Session{AbsoluteExpiresAt: abs, IdleExpiresAt: time.Now()}
	s.RefreshIdle(5 * time.Minute)
	if !s.IdleExpiresAt.Before(abs) {
		t.Error("idle should be before absolute")
	}
}

func TestSession_RefreshIdle_CappedAtAbsolute(t *testing.T) {
	abs := time.Now().Add(time.Minute)
	s := &domain.Session{AbsoluteExpiresAt: abs, IdleExpiresAt: time.Now()}
	s.RefreshIdle(time.Hour) // would overshoot absolute
	if !s.IdleExpiresAt.Equal(abs) {
		t.Errorf("expected idle capped at abs, got %v vs %v", s.IdleExpiresAt, abs)
	}
}
