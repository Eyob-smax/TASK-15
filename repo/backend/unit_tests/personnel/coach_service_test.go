package personnel_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
)

func TestCoachCreate_Success(t *testing.T) {
	repo := newMockCoachRepo()
	svc := application.NewCoachService(repo)

	userID := uuid.New()
	locationID := uuid.New()

	coach, err := svc.Create(context.Background(), &domain.Coach{
		UserID:         userID,
		LocationID:     locationID,
		Specialization: "Strength",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if coach.ID == uuid.Nil {
		t.Error("expected ID assigned")
	}
	if !coach.IsActive {
		t.Error("expected IsActive=true")
	}
	if coach.Specialization != "Strength" {
		t.Errorf("specialization not preserved: got %q", coach.Specialization)
	}
}

func TestCoachCreate_MissingUserID_Validation(t *testing.T) {
	repo := newMockCoachRepo()
	svc := application.NewCoachService(repo)

	_, err := svc.Create(context.Background(), &domain.Coach{LocationID: uuid.New()})
	var vErr *domain.ErrValidation
	if !errors.As(err, &vErr) || vErr.Field != "user_id" {
		t.Fatalf("expected validation on user_id, got %v", err)
	}
}

func TestCoachCreate_MissingLocationID_Validation(t *testing.T) {
	repo := newMockCoachRepo()
	svc := application.NewCoachService(repo)

	_, err := svc.Create(context.Background(), &domain.Coach{UserID: uuid.New()})
	var vErr *domain.ErrValidation
	if !errors.As(err, &vErr) || vErr.Field != "location_id" {
		t.Fatalf("expected validation on location_id, got %v", err)
	}
}

func TestCoachGetByID_Found(t *testing.T) {
	repo := newMockCoachRepo()
	svc := application.NewCoachService(repo)

	id := uuid.New()
	repo.coaches[id] = &domain.Coach{ID: id, UserID: uuid.New(), LocationID: uuid.New()}

	got, err := svc.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.ID != id {
		t.Errorf("expected ID=%v, got %v", id, got.ID)
	}
}

func TestCoachGetByID_NotFound(t *testing.T) {
	repo := newMockCoachRepo()
	svc := application.NewCoachService(repo)

	_, err := svc.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCoachList_FilteredByLocation(t *testing.T) {
	repo := newMockCoachRepo()
	svc := application.NewCoachService(repo)

	locA := uuid.New()
	locB := uuid.New()
	repo.coaches[uuid.New()] = &domain.Coach{ID: uuid.New(), UserID: uuid.New(), LocationID: locA}
	repo.coaches[uuid.New()] = &domain.Coach{ID: uuid.New(), UserID: uuid.New(), LocationID: locB}

	rows, total, err := svc.List(context.Background(), &locA, 1, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 || len(rows) != 1 || rows[0].LocationID != locA {
		t.Errorf("expected filter by locA to return 1 row, got total=%d rows=%+v", total, rows)
	}
}

func TestCoachGetByIDForActor_NilActor_Forbidden(t *testing.T) {
	repo := newMockCoachRepo()
	svc := application.NewCoachService(repo)

	id := uuid.New()
	repo.coaches[id] = &domain.Coach{ID: id, UserID: uuid.New(), LocationID: uuid.New()}

	_, err := svc.GetByIDForActor(context.Background(), nil, id)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden for nil actor, got %v", err)
	}
}

func TestCoachGetByIDForActor_Admin_AllowedAcrossLocations(t *testing.T) {
	repo := newMockCoachRepo()
	svc := application.NewCoachService(repo)

	id := uuid.New()
	repo.coaches[id] = &domain.Coach{ID: id, UserID: uuid.New(), LocationID: uuid.New()}
	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleAdministrator}
	got, err := svc.GetByIDForActor(context.Background(), actor, id)
	if err != nil {
		t.Fatalf("expected admin to read any coach, got %v", err)
	}
	if got.ID != id {
		t.Error("unexpected coach returned")
	}
}

func TestCoachGetByIDForActor_OpsManagerSameLocation_Allowed(t *testing.T) {
	repo := newMockCoachRepo()
	svc := application.NewCoachService(repo)

	locationID := uuid.New()
	id := uuid.New()
	repo.coaches[id] = &domain.Coach{ID: id, UserID: uuid.New(), LocationID: locationID}

	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleOperationsManager, LocationID: &locationID}
	got, err := svc.GetByIDForActor(context.Background(), actor, id)
	if err != nil {
		t.Fatalf("expected allowed, got %v", err)
	}
	if got.LocationID != locationID {
		t.Error("expected same-location coach")
	}
}

func TestCoachGetByIDForActor_OpsManagerNoLocation_Forbidden(t *testing.T) {
	repo := newMockCoachRepo()
	svc := application.NewCoachService(repo)
	id := uuid.New()
	repo.coaches[id] = &domain.Coach{ID: id, UserID: uuid.New(), LocationID: uuid.New()}
	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleOperationsManager}
	_, err := svc.GetByIDForActor(context.Background(), actor, id)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestCoachGetByIDForActor_NotFound_PropagatesError(t *testing.T) {
	repo := newMockCoachRepo()
	svc := application.NewCoachService(repo)
	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleAdministrator}
	_, err := svc.GetByIDForActor(context.Background(), actor, uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCoachListForActor_NilActor_Forbidden(t *testing.T) {
	repo := newMockCoachRepo()
	svc := application.NewCoachService(repo)
	_, _, err := svc.ListForActor(context.Background(), nil, nil, 1, 10)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestCoachListForActor_AdminPassesRequestedLocation(t *testing.T) {
	repo := newMockCoachRepo()
	svc := application.NewCoachService(repo)

	locA := uuid.New()
	locB := uuid.New()
	repo.coaches[uuid.New()] = &domain.Coach{ID: uuid.New(), UserID: uuid.New(), LocationID: locA}
	repo.coaches[uuid.New()] = &domain.Coach{ID: uuid.New(), UserID: uuid.New(), LocationID: locB}

	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleAdministrator}
	rows, total, err := svc.ListForActor(context.Background(), actor, &locA, 1, 10)
	if err != nil {
		t.Fatalf("ListForActor failed: %v", err)
	}
	if total != 1 || len(rows) != 1 {
		t.Errorf("expected 1 row, got total=%d rows=%+v", total, rows)
	}
}

func TestCoachListForActor_OpsManagerNoLocation_Forbidden(t *testing.T) {
	repo := newMockCoachRepo()
	svc := application.NewCoachService(repo)
	actor := &domain.User{ID: uuid.New(), Role: domain.UserRoleOperationsManager}
	_, _, err := svc.ListForActor(context.Background(), actor, nil, 1, 10)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
