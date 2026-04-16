package personnel_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
)

func TestMemberCreate_Success(t *testing.T) {
	repo := newMockMemberRepo()
	svc := application.NewMemberService(repo)

	userID := uuid.New()
	locationID := uuid.New()

	member, err := svc.Create(context.Background(), &domain.Member{
		UserID:     userID,
		LocationID: locationID,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if member.ID == uuid.Nil {
		t.Error("expected ID to be assigned")
	}
	if member.MembershipStatus != domain.MembershipStatusActive {
		t.Errorf("expected status=active, got %v", member.MembershipStatus)
	}
	if member.JoinedAt.IsZero() {
		t.Error("expected JoinedAt to be set")
	}
	if _, ok := repo.members[member.ID]; !ok {
		t.Error("expected member to be persisted")
	}
}

func TestMemberCreate_MissingUserID_Validation(t *testing.T) {
	repo := newMockMemberRepo()
	svc := application.NewMemberService(repo)

	_, err := svc.Create(context.Background(), &domain.Member{
		LocationID: uuid.New(),
	})
	var vErr *domain.ErrValidation
	if !errors.As(err, &vErr) || vErr.Field != "user_id" {
		t.Fatalf("expected ErrValidation on user_id, got %v", err)
	}
}

func TestMemberCreate_MissingLocationID_Validation(t *testing.T) {
	repo := newMockMemberRepo()
	svc := application.NewMemberService(repo)

	_, err := svc.Create(context.Background(), &domain.Member{
		UserID: uuid.New(),
	})
	var vErr *domain.ErrValidation
	if !errors.As(err, &vErr) || vErr.Field != "location_id" {
		t.Fatalf("expected ErrValidation on location_id, got %v", err)
	}
}

func TestMemberGetByID_Found(t *testing.T) {
	repo := newMockMemberRepo()
	svc := application.NewMemberService(repo)

	id := uuid.New()
	repo.members[id] = &domain.Member{ID: id, UserID: uuid.New(), LocationID: uuid.New()}

	got, err := svc.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.ID != id {
		t.Errorf("expected ID=%v, got %v", id, got.ID)
	}
}

func TestMemberGetByID_NotFound(t *testing.T) {
	repo := newMockMemberRepo()
	svc := application.NewMemberService(repo)

	_, err := svc.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestMemberList_Paginated(t *testing.T) {
	repo := newMockMemberRepo()
	svc := application.NewMemberService(repo)

	locationID := uuid.New()
	for i := 0; i < 3; i++ {
		id := uuid.New()
		repo.members[id] = &domain.Member{ID: id, UserID: uuid.New(), LocationID: locationID}
	}
	rows, total, err := svc.List(context.Background(), &locationID, 1, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 3 || len(rows) != 3 {
		t.Errorf("expected 3 members, got total=%d rows=%d", total, len(rows))
	}
}

func TestMemberGetByIDForActor_NilActor_Forbidden(t *testing.T) {
	repo := newMockMemberRepo()
	svc := application.NewMemberService(repo)

	id := uuid.New()
	repo.members[id] = &domain.Member{ID: id, UserID: uuid.New(), LocationID: uuid.New()}

	_, err := svc.GetByIDForActor(context.Background(), nil, id)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestMemberGetByIDForActor_AdministratorAllowedAcrossLocations(t *testing.T) {
	repo := newMockMemberRepo()
	svc := application.NewMemberService(repo)

	id := uuid.New()
	repo.members[id] = &domain.Member{ID: id, UserID: uuid.New(), LocationID: uuid.New()}

	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleAdministrator}
	got, err := svc.GetByIDForActor(context.Background(), actor, id)
	if err != nil {
		t.Fatalf("GetByIDForActor failed: %v", err)
	}
	if got.ID != id {
		t.Error("expected to return the member for admin")
	}
}

func TestMemberGetByIDForActor_OperationsManagerSameLocation_Allowed(t *testing.T) {
	repo := newMockMemberRepo()
	svc := application.NewMemberService(repo)

	locationID := uuid.New()
	id := uuid.New()
	repo.members[id] = &domain.Member{ID: id, UserID: uuid.New(), LocationID: locationID}

	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleOperationsManager, LocationID: &locationID}
	got, err := svc.GetByIDForActor(context.Background(), actor, id)
	if err != nil {
		t.Fatalf("GetByIDForActor same-location failed: %v", err)
	}
	if got.ID != id {
		t.Error("expected to return the member for same-location ops manager")
	}
}

func TestMemberGetByIDForActor_OperationsManagerWithoutLocation_Forbidden(t *testing.T) {
	repo := newMockMemberRepo()
	svc := application.NewMemberService(repo)

	id := uuid.New()
	repo.members[id] = &domain.Member{ID: id, UserID: uuid.New(), LocationID: uuid.New()}

	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleOperationsManager}
	_, err := svc.GetByIDForActor(context.Background(), actor, id)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for ops manager with no location, got %v", err)
	}
}

func TestMemberGetByIDForActor_NotFound_PropagatesError(t *testing.T) {
	repo := newMockMemberRepo()
	svc := application.NewMemberService(repo)

	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleAdministrator}
	_, err := svc.GetByIDForActor(context.Background(), actor, uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestMemberListForActor_NilActor_Forbidden(t *testing.T) {
	repo := newMockMemberRepo()
	svc := application.NewMemberService(repo)
	_, _, err := svc.ListForActor(context.Background(), nil, nil, 1, 10)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for nil actor, got %v", err)
	}
}

func TestMemberListForActor_AdministratorPassesRequestedLocation(t *testing.T) {
	repo := newMockMemberRepo()
	svc := application.NewMemberService(repo)

	locA := uuid.New()
	locB := uuid.New()
	repo.members[uuid.New()] = &domain.Member{ID: uuid.New(), UserID: uuid.New(), LocationID: locA}
	repo.members[uuid.New()] = &domain.Member{ID: uuid.New(), UserID: uuid.New(), LocationID: locB}

	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleAdministrator}
	rows, total, err := svc.ListForActor(context.Background(), actor, &locA, 1, 10)
	if err != nil {
		t.Fatalf("ListForActor failed: %v", err)
	}
	if total != 1 || len(rows) != 1 || rows[0].LocationID != locA {
		t.Fatalf("expected admin list to honor requested location, got total=%d rows=%+v", total, rows)
	}
}

func TestMemberListForActor_OpsManagerNoLocation_Forbidden(t *testing.T) {
	repo := newMockMemberRepo()
	svc := application.NewMemberService(repo)
	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleOperationsManager}
	_, _, err := svc.ListForActor(context.Background(), actor, nil, 1, 10)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for ops manager with no location, got %v", err)
	}
}
