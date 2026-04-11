package admin_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
)

// --- Mock repositories ---

type mockUserRepo struct {
	users      map[uuid.UUID]*domain.User
	byEmail    map[string]*domain.User
	createErr  error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:   make(map[uuid.UUID]*domain.User),
		byEmail: make(map[string]*domain.User),
	}
}

func (m *mockUserRepo) Create(_ context.Context, user *domain.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.users[user.ID] = user
	m.byEmail[user.Email] = user
	return nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return user, nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	user, ok := m.byEmail[email]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return user, nil
}

func (m *mockUserRepo) Update(_ context.Context, user *domain.User) error {
	if _, ok := m.users[user.ID]; !ok {
		return domain.ErrNotFound
	}
	m.users[user.ID] = user
	m.byEmail[user.Email] = user
	return nil
}

func (m *mockUserRepo) List(_ context.Context, _, _ int) ([]domain.User, int, error) {
	list := make([]domain.User, 0, len(m.users))
	for _, v := range m.users {
		list = append(list, *v)
	}
	return list, len(list), nil
}

func (m *mockUserRepo) ListByRole(_ context.Context, _ domain.UserRole, _, _ int) ([]domain.User, int, error) {
	return nil, 0, nil
}

// --- Mock audit service ---

type mockAuditService struct{}

func (m *mockAuditService) Log(_ context.Context, _, _ string, _, _ uuid.UUID, _ map[string]interface{}) error {
	return nil
}

func (m *mockAuditService) List(_ context.Context, _ string, _ *uuid.UUID, _, _ int) ([]domain.AuditEvent, int, error) {
	return nil, 0, nil
}

func (m *mockAuditService) ListByEventTypes(_ context.Context, _ []string, _, _ int) ([]domain.AuditEvent, int, error) {
	return nil, 0, nil
}

// --- Helper ---

func newUserService(repo *mockUserRepo) application.UserService {
	return application.NewUserService(repo, &mockAuditService{})
}

// --- Tests ---

func TestUserCreate_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := newUserService(repo)

	user, err := svc.Create(context.Background(), "alice@example.com", "password123", domain.UserRoleAdministrator, "Alice", nil)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user from Create")
	}
	if user.PasswordHash == "" {
		t.Error("expected PasswordHash to be set")
	}
	if user.Email != "alice@example.com" {
		t.Errorf("expected email=alice@example.com, got %v", user.Email)
	}
	// Confirm it was stored in repo.
	if _, ok := repo.users[user.ID]; !ok {
		t.Error("expected user to be persisted in repo")
	}
}

func TestUserCreate_RepoError(t *testing.T) {
	repo := newMockUserRepo()
	repo.createErr = errors.New("database connection lost")
	svc := newUserService(repo)

	_, err := svc.Create(context.Background(), "bob@example.com", "pass", domain.UserRoleCoach, "Bob", nil)
	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}
}

func TestUserGet_Found(t *testing.T) {
	repo := newMockUserRepo()
	svc := newUserService(repo)

	id := uuid.New()
	repo.users[id] = &domain.User{
		ID:        id,
		Email:     "carol@example.com",
		Role:      domain.UserRoleMember,
		Status:    domain.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	user, err := svc.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if user.ID != id {
		t.Errorf("expected ID=%v, got %v", id, user.ID)
	}
}

func TestUserList_Paginated(t *testing.T) {
	repo := newMockUserRepo()
	svc := newUserService(repo)

	// Pre-seed two users.
	for i := 0; i < 2; i++ {
		id := uuid.New()
		repo.users[id] = &domain.User{
			ID:        id,
			Email:     uuid.New().String() + "@example.com",
			Role:      domain.UserRoleMember,
			Status:    domain.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	users, total, err := svc.List(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total=2, got %d", total)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestUserDeactivate_SetsInactive(t *testing.T) {
	repo := newMockUserRepo()
	svc := newUserService(repo)

	id := uuid.New()
	repo.users[id] = &domain.User{
		ID:        id,
		Email:     "dave@example.com",
		Role:      domain.UserRoleMember,
		Status:    domain.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := svc.Deactivate(context.Background(), id); err != nil {
		t.Fatalf("Deactivate failed: %v", err)
	}

	updated := repo.users[id]
	if updated.Status != domain.UserStatusInactive {
		t.Errorf("expected status=inactive, got %v", updated.Status)
	}
}
