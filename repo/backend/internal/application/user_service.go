package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/security"
	"fitcommerce/internal/store"
)

// UserServiceImpl implements UserService.
type UserServiceImpl struct {
	userRepo store.UserRepository
	auditSvc AuditService
}

// NewUserService creates a UserServiceImpl backed by the given repository.
func NewUserService(userRepo store.UserRepository, auditSvc AuditService) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo: userRepo,
		auditSvc: auditSvc,
	}
}

// Create creates a new user account with a hashed password.
func (s *UserServiceImpl) Create(
	ctx context.Context,
	email, password string,
	role domain.UserRole,
	displayName string,
	locationID *uuid.UUID,
) (*domain.User, error) {
	hash, salt, err := security.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("user_service.Create hash password: %w", err)
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hash,
		Salt:         salt,
		Role:         role,
		DisplayName:  displayName,
		LocationID:   locationID,
		Status:       domain.UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	_ = s.auditSvc.Log(ctx, "user.created", "user", user.ID, user.ID, map[string]interface{}{
		"email": user.Email,
		"role":  string(user.Role),
	})

	return user, nil
}

// Get retrieves a user by ID.
func (s *UserServiceImpl) Get(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// List returns a paginated list of users.
func (s *UserServiceImpl) List(ctx context.Context, page, pageSize int) ([]domain.User, int, error) {
	return s.userRepo.List(ctx, page, pageSize)
}

// Update persists changes to an existing user record.
func (s *UserServiceImpl) Update(ctx context.Context, user *domain.User) error {
	if _, err := s.userRepo.GetByID(ctx, user.ID); err != nil {
		return err
	}

	user.UpdatedAt = time.Now().UTC()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	_ = s.auditSvc.Log(ctx, "user.updated", "user", user.ID, user.ID, map[string]interface{}{})

	return nil
}

// Deactivate sets the user's status to inactive.
func (s *UserServiceImpl) Deactivate(ctx context.Context, id uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	user.Status = domain.UserStatusInactive
	user.UpdatedAt = time.Now().UTC()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	if err := s.auditSvc.Log(ctx, "user.deactivated", "user", user.ID, domain.SystemActorID, map[string]interface{}{}); err != nil {
		slog.Default().Warn("audit log failed", "event", "user.deactivated", "error", err)
	}

	return nil
}

// Compile-time interface assertion.
var _ UserService = (*UserServiceImpl)(nil)
