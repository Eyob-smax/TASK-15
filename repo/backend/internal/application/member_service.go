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

// MemberServiceImpl implements MemberService.
type MemberServiceImpl struct {
	repo store.MemberRepository
}

// NewMemberService creates a MemberServiceImpl backed by the given repository.
func NewMemberService(repo store.MemberRepository) *MemberServiceImpl {
	return &MemberServiceImpl{repo: repo}
}

// Create validates and persists a new member.
func (s *MemberServiceImpl) Create(ctx context.Context, member *domain.Member) (*domain.Member, error) {
	if member.UserID == uuid.Nil {
		return nil, &domain.ErrValidation{Field: "user_id", Message: "user_id is required"}
	}
	if member.LocationID == uuid.Nil {
		return nil, &domain.ErrValidation{Field: "location_id", Message: "location_id is required"}
	}
	now := time.Now().UTC()
	member.ID = uuid.New()
	member.MembershipStatus = domain.MembershipStatusActive
	member.JoinedAt = now
	member.CreatedAt = now
	member.UpdatedAt = now
	if err := s.repo.Create(ctx, member); err != nil {
		return nil, fmt.Errorf("member_service.Create: %w", err)
	}
	return member, nil
}

// GetByID retrieves a member by ID.
func (s *MemberServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.Member, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByIDForActor retrieves a member and enforces actor location scope.
func (s *MemberServiceImpl) GetByIDForActor(ctx context.Context, actor *domain.User, id uuid.UUID) (*domain.Member, error) {
	member, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if actor == nil || !security.HasPermission(actor.Role, security.ActionViewMembers) {
		return nil, domain.ErrForbidden
	}
	if actor.Role == domain.UserRoleAdministrator {
		return member, nil
	}
	if actor.LocationID == nil || member.LocationID != *actor.LocationID {
		return nil, domain.ErrForbidden
	}
	return member, nil
}

// List returns a paginated list of members, optionally filtered by location.
func (s *MemberServiceImpl) List(ctx context.Context, locationID *uuid.UUID, page, pageSize int) ([]domain.Member, int, error) {
	return s.repo.List(ctx, locationID, page, pageSize)
}

// ListForActor returns a paginated list of members using actor-aware location scoping.
func (s *MemberServiceImpl) ListForActor(ctx context.Context, actor *domain.User, requestedLocationID *uuid.UUID, page, pageSize int) ([]domain.Member, int, error) {
	if actor == nil || !security.HasPermission(actor.Role, security.ActionViewMembers) {
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
var _ MemberService = (*MemberServiceImpl)(nil)
