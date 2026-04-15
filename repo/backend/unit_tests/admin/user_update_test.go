package admin_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
)

func TestUserUpdate_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := newUserService(repo)

	id := uuid.New()
	before := time.Now().Add(-time.Hour)
	repo.users[id] = &domain.User{
		ID:          id,
		Email:       "alice@example.com",
		Role:        domain.UserRoleMember,
		Status:      domain.UserStatusActive,
		DisplayName: "Old Alice",
		UpdatedAt:   before,
	}
	repo.byEmail["alice@example.com"] = repo.users[id]

	update := &domain.User{
		ID:          id,
		Email:       "alice@example.com",
		Role:        domain.UserRoleAdministrator,
		Status:      domain.UserStatusActive,
		DisplayName: "Alice the Admin",
	}
	if err := svc.Update(context.Background(), update); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	stored := repo.users[id]
	if stored.DisplayName != "Alice the Admin" {
		t.Errorf("display name not applied: %q", stored.DisplayName)
	}
	if stored.Role != domain.UserRoleAdministrator {
		t.Errorf("role not applied: %v", stored.Role)
	}
	if !stored.UpdatedAt.After(before) {
		t.Error("expected UpdatedAt to advance")
	}
}

func TestUserUpdate_NotFound_ReturnsError(t *testing.T) {
	repo := newMockUserRepo()
	svc := newUserService(repo)

	err := svc.Update(context.Background(), &domain.User{ID: uuid.New(), Email: "ghost@example.com"})
	if err == nil {
		t.Fatal("expected error when user does not exist")
	}
}
