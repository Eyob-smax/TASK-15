package catalog_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
)

// --- InventoryRepository mock ---

type invServiceInvRepo struct {
	snapshots   []domain.InventorySnapshot
	adjustments []domain.InventoryAdjustment
}

func (m *invServiceInvRepo) CreateSnapshot(_ context.Context, s *domain.InventorySnapshot) error {
	m.snapshots = append(m.snapshots, *s)
	return nil
}
func (m *invServiceInvRepo) ListSnapshots(_ context.Context, itemID *uuid.UUID, _ *uuid.UUID) ([]domain.InventorySnapshot, error) {
	if itemID == nil {
		return m.snapshots, nil
	}
	var out []domain.InventorySnapshot
	for _, s := range m.snapshots {
		if s.ItemID == *itemID {
			out = append(out, s)
		}
	}
	return out, nil
}
func (m *invServiceInvRepo) CreateAdjustment(_ context.Context, a *domain.InventoryAdjustment) error {
	m.adjustments = append(m.adjustments, *a)
	return nil
}
func (m *invServiceInvRepo) ListAdjustments(_ context.Context, _ *uuid.UUID, _, _ int) ([]domain.InventoryAdjustment, int, error) {
	return m.adjustments, len(m.adjustments), nil
}

// --- WarehouseBinRepository mock ---

type mockBinRepo struct {
	bins   map[uuid.UUID]*domain.WarehouseBin
	getErr error
}

func newMockBinRepo() *mockBinRepo {
	return &mockBinRepo{bins: make(map[uuid.UUID]*domain.WarehouseBin)}
}

func (m *mockBinRepo) Create(_ context.Context, b *domain.WarehouseBin) error {
	m.bins[b.ID] = b
	return nil
}
func (m *mockBinRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.WarehouseBin, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	b, ok := m.bins[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return b, nil
}
func (m *mockBinRepo) List(_ context.Context, locationID *uuid.UUID, _, _ int) ([]domain.WarehouseBin, int, error) {
	out := make([]domain.WarehouseBin, 0, len(m.bins))
	for _, b := range m.bins {
		if locationID != nil && b.LocationID != *locationID {
			continue
		}
		out = append(out, *b)
	}
	return out, len(out), nil
}

func TestInventoryService_CreateWarehouseBin_SetsIDAndTimestamps(t *testing.T) {
	invRepo := &invServiceInvRepo{}
	binRepo := newMockBinRepo()
	itemRepo := newMockItemRepo()
	svc := application.NewInventoryService(invRepo, binRepo, itemRepo, &mockAuditService{}, nil)

	bin, err := svc.CreateWarehouseBin(context.Background(), &domain.WarehouseBin{
		LocationID: uuid.New(),
		Name:       "A1",
	})
	if err != nil {
		t.Fatalf("CreateWarehouseBin failed: %v", err)
	}
	if bin.ID == uuid.Nil {
		t.Error("expected ID assigned")
	}
	if bin.CreatedAt.IsZero() || bin.UpdatedAt.IsZero() {
		t.Error("expected timestamps set")
	}
}

func TestInventoryService_GetWarehouseBin_Found(t *testing.T) {
	invRepo := &invServiceInvRepo{}
	binRepo := newMockBinRepo()
	itemRepo := newMockItemRepo()
	svc := application.NewInventoryService(invRepo, binRepo, itemRepo, &mockAuditService{}, nil)

	id := uuid.New()
	binRepo.bins[id] = &domain.WarehouseBin{ID: id, Name: "A1"}

	got, err := svc.GetWarehouseBin(context.Background(), id)
	if err != nil {
		t.Fatalf("GetWarehouseBin failed: %v", err)
	}
	if got.ID != id {
		t.Errorf("expected %v got %v", id, got.ID)
	}
}

func TestInventoryService_GetWarehouseBin_NotFound(t *testing.T) {
	invRepo := &invServiceInvRepo{}
	binRepo := newMockBinRepo()
	itemRepo := newMockItemRepo()
	svc := application.NewInventoryService(invRepo, binRepo, itemRepo, &mockAuditService{}, nil)

	_, err := svc.GetWarehouseBin(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestInventoryService_ListWarehouseBins_FilterByLocation(t *testing.T) {
	invRepo := &invServiceInvRepo{}
	binRepo := newMockBinRepo()
	itemRepo := newMockItemRepo()
	svc := application.NewInventoryService(invRepo, binRepo, itemRepo, &mockAuditService{}, nil)

	locA := uuid.New()
	locB := uuid.New()
	for i := 0; i < 2; i++ {
		id := uuid.New()
		binRepo.bins[id] = &domain.WarehouseBin{ID: id, LocationID: locA}
	}
	id := uuid.New()
	binRepo.bins[id] = &domain.WarehouseBin{ID: id, LocationID: locB}

	rows, total, err := svc.ListWarehouseBins(context.Background(), &locA, 1, 10)
	if err != nil {
		t.Fatalf("ListWarehouseBins failed: %v", err)
	}
	if total != 2 || len(rows) != 2 {
		t.Errorf("expected 2 bins at locA, got total=%d rows=%d", total, len(rows))
	}
}
