package personnel_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
)

type mockLocationRepo struct {
	locations map[uuid.UUID]*domain.Location
	createErr error
}

func newMockLocationRepo() *mockLocationRepo {
	return &mockLocationRepo{locations: make(map[uuid.UUID]*domain.Location)}
}

func (m *mockLocationRepo) Create(_ context.Context, loc *domain.Location) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.locations[loc.ID] = loc
	return nil
}
func (m *mockLocationRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Location, error) {
	loc, ok := m.locations[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return loc, nil
}
func (m *mockLocationRepo) List(_ context.Context, _, _ int) ([]domain.Location, int, error) {
	list := make([]domain.Location, 0, len(m.locations))
	for _, l := range m.locations {
		list = append(list, *l)
	}
	return list, len(list), nil
}

func TestLocationCreate_Success(t *testing.T) {
	repo := newMockLocationRepo()
	svc := application.NewLocationService(repo)

	loc, err := svc.Create(context.Background(), &domain.Location{Name: "HQ", Timezone: "UTC"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if loc.ID == uuid.Nil {
		t.Error("expected ID assigned")
	}
	if !loc.IsActive {
		t.Error("expected new location to be IsActive=true")
	}
	if loc.CreatedAt.IsZero() || loc.UpdatedAt.IsZero() {
		t.Error("expected CreatedAt/UpdatedAt set")
	}
}

func TestLocationCreate_MissingName_Validation(t *testing.T) {
	repo := newMockLocationRepo()
	svc := application.NewLocationService(repo)

	_, err := svc.Create(context.Background(), &domain.Location{})
	var vErr *domain.ErrValidation
	if !errors.As(err, &vErr) || vErr.Field != "name" {
		t.Fatalf("expected validation on name, got %v", err)
	}
}

func TestLocationCreate_RepoError_Wrapped(t *testing.T) {
	repo := newMockLocationRepo()
	repo.createErr = errors.New("boom")
	svc := application.NewLocationService(repo)

	_, err := svc.Create(context.Background(), &domain.Location{Name: "X"})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestLocationGetByID_Found(t *testing.T) {
	repo := newMockLocationRepo()
	svc := application.NewLocationService(repo)

	id := uuid.New()
	repo.locations[id] = &domain.Location{ID: id, Name: "Loc"}

	got, err := svc.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.ID != id {
		t.Errorf("expected %v got %v", id, got.ID)
	}
}

func TestLocationGetByID_NotFound(t *testing.T) {
	repo := newMockLocationRepo()
	svc := application.NewLocationService(repo)

	_, err := svc.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestLocationList(t *testing.T) {
	repo := newMockLocationRepo()
	svc := application.NewLocationService(repo)

	for i := 0; i < 2; i++ {
		id := uuid.New()
		repo.locations[id] = &domain.Location{ID: id, Name: "Loc"}
	}
	rows, total, err := svc.List(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 2 || len(rows) != 2 {
		t.Errorf("expected 2 rows, got total=%d rows=%d", total, len(rows))
	}
}
