package personnel_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
)

type mockMemberRepo struct {
	members map[uuid.UUID]*domain.Member
}

func newMockMemberRepo() *mockMemberRepo {
	return &mockMemberRepo{members: make(map[uuid.UUID]*domain.Member)}
}

func (m *mockMemberRepo) Create(_ context.Context, member *domain.Member) error {
	m.members[member.ID] = member
	return nil
}

func (m *mockMemberRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Member, error) {
	member, ok := m.members[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return member, nil
}

func (m *mockMemberRepo) GetByUserID(_ context.Context, userID uuid.UUID) (*domain.Member, error) {
	for _, member := range m.members {
		if member.UserID == userID {
			return member, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockMemberRepo) List(_ context.Context, locationID *uuid.UUID, _, _ int) ([]domain.Member, int, error) {
	var rows []domain.Member
	for _, member := range m.members {
		if locationID != nil && member.LocationID != *locationID {
			continue
		}
		rows = append(rows, *member)
	}
	return rows, len(rows), nil
}

func (m *mockMemberRepo) CountByPeriod(_ context.Context, _ *uuid.UUID, _, _ time.Time) (int, error) {
	return 0, nil
}

type mockCoachRepo struct {
	coaches map[uuid.UUID]*domain.Coach
}

func newMockCoachRepo() *mockCoachRepo {
	return &mockCoachRepo{coaches: make(map[uuid.UUID]*domain.Coach)}
}

func (m *mockCoachRepo) Create(_ context.Context, coach *domain.Coach) error {
	m.coaches[coach.ID] = coach
	return nil
}

func (m *mockCoachRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Coach, error) {
	coach, ok := m.coaches[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return coach, nil
}

func (m *mockCoachRepo) GetByUserID(_ context.Context, userID uuid.UUID) (*domain.Coach, error) {
	for _, coach := range m.coaches {
		if coach.UserID == userID {
			return coach, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockCoachRepo) List(_ context.Context, locationID *uuid.UUID, _, _ int) ([]domain.Coach, int, error) {
	var rows []domain.Coach
	for _, coach := range m.coaches {
		if locationID != nil && coach.LocationID != *locationID {
			continue
		}
		rows = append(rows, *coach)
	}
	return rows, len(rows), nil
}

func TestMemberService_ListForActor_ForcesAssignedLocation(t *testing.T) {
	repo := newMockMemberRepo()
	service := application.NewMemberService(repo)

	locationA := uuid.New()
	locationB := uuid.New()
	repo.members[uuid.New()] = &domain.Member{ID: uuid.New(), UserID: uuid.New(), LocationID: locationA}
	repo.members[uuid.New()] = &domain.Member{ID: uuid.New(), UserID: uuid.New(), LocationID: locationB}

	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleOperationsManager, LocationID: &locationA}
	rows, total, err := service.ListForActor(context.Background(), actor, &locationB, 1, 20)
	if err != nil {
		t.Fatalf("ListForActor failed: %v", err)
	}
	if total != 1 || len(rows) != 1 || rows[0].LocationID != locationA {
		t.Fatalf("expected actor scope to force location A, got total=%d rows=%#v", total, rows)
	}
}

func TestMemberService_GetByIDForActor_RejectsRoleWithoutPermission(t *testing.T) {
	repo := newMockMemberRepo()
	service := application.NewMemberService(repo)

	locationID := uuid.New()
	memberID := uuid.New()
	repo.members[memberID] = &domain.Member{ID: memberID, UserID: uuid.New(), LocationID: locationID}

	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleCoach, LocationID: &locationID}
	if _, err := service.GetByIDForActor(context.Background(), actor, memberID); err != domain.ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestCoachService_GetByIDForActor_RejectsCrossLocation(t *testing.T) {
	repo := newMockCoachRepo()
	service := application.NewCoachService(repo)

	locationA := uuid.New()
	locationB := uuid.New()
	coachID := uuid.New()
	repo.coaches[coachID] = &domain.Coach{ID: coachID, UserID: uuid.New(), LocationID: locationB}

	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleOperationsManager, LocationID: &locationA}
	if _, err := service.GetByIDForActor(context.Background(), actor, coachID); err != domain.ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestCoachService_ListForActor_RejectsRoleWithoutPermission(t *testing.T) {
	repo := newMockCoachRepo()
	service := application.NewCoachService(repo)

	locationID := uuid.New()
	repo.coaches[uuid.New()] = &domain.Coach{ID: uuid.New(), UserID: uuid.New(), LocationID: locationID}

	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleMember, LocationID: &locationID}
	if _, _, err := service.ListForActor(context.Background(), actor, &locationID, 1, 20); err != domain.ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
