package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/security"
	"fitcommerce/internal/store"
)

// CoachServiceImpl implements CoachService.
type CoachServiceImpl struct {
	repo store.CoachRepository
}

// NewCoachService creates a CoachServiceImpl backed by the given repository.
func NewCoachService(repo store.CoachRepository) *CoachServiceImpl {
	return &CoachServiceImpl{repo: repo}
}

// Create validates and persists a new coach.
func (s *CoachServiceImpl) Create(ctx context.Context, coach *domain.Coach) (*domain.Coach, error) {
	if coach.UserID == uuid.Nil {
		return nil, &domain.ErrValidation{Field: "user_id", Message: "user_id is required"}
	}
	if coach.LocationID == uuid.Nil {
		return nil, &domain.ErrValidation{Field: "location_id", Message: "location_id is required"}
	}
	now := time.Now().UTC()
	coach.ID = uuid.New()
	coach.IsActive = true
	coach.CreatedAt = now
	coach.UpdatedAt = now
	if err := s.repo.Create(ctx, coach); err != nil {
		return nil, fmt.Errorf("coach_service.Create: %w", err)
	}
	return coach, nil
}

// GetByID retrieves a coach by ID.
func (s *CoachServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.Coach, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByIDForActor retrieves a coach and enforces actor location scope.
func (s *CoachServiceImpl) GetByIDForActor(ctx context.Context, actor *domain.User, id uuid.UUID) (*domain.Coach, error) {
	coach, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if actor == nil || !security.HasPermission(actor.Role, security.ActionViewCoaches) {
		return nil, domain.ErrForbidden
	}
	if actor.Role == domain.UserRoleAdministrator {
		return coach, nil
	}
	if actor.LocationID == nil || coach.LocationID != *actor.LocationID {
		return nil, domain.ErrForbidden
	}
	return coach, nil
}

// List returns a paginated list of coaches, optionally filtered by location.
func (s *CoachServiceImpl) List(ctx context.Context, locationID *uuid.UUID, page, pageSize int) ([]domain.Coach, int, error) {
	return s.repo.List(ctx, locationID, page, pageSize)
}

// ListForActor returns a paginated list of coaches using actor-aware location scoping.
func (s *CoachServiceImpl) ListForActor(ctx context.Context, actor *domain.User, requestedLocationID *uuid.UUID, page, pageSize int) ([]domain.Coach, int, error) {
	if actor == nil || !security.HasPermission(actor.Role, security.ActionViewCoaches) {
		return nil, 0, domain.ErrForbidden
	}

	if actor.Role == domain.UserRoleAdministrator {
		return s.repo.List(ctx, requestedLocationID, page, pageSize)
	}
	if actor.LocationID == nil {
		return nil, 0, domain.ErrForbidden
	}

	locationID := requestedLocationID
	locationID = actor.LocationID

	return s.repo.List(ctx, locationID, page, pageSize)
}

// Compile-time interface assertion.
var _ CoachService = (*CoachServiceImpl)(nil)
