package procurement_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
)

type mockSupplierRepo struct {
	suppliers map[uuid.UUID]*domain.Supplier
	createErr error
	updateErr error
}

func newMockSupplierRepo() *mockSupplierRepo {
	return &mockSupplierRepo{suppliers: make(map[uuid.UUID]*domain.Supplier)}
}

func (m *mockSupplierRepo) Create(_ context.Context, s *domain.Supplier) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.suppliers[s.ID] = s
	return nil
}

func (m *mockSupplierRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Supplier, error) {
	s, ok := m.suppliers[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return s, nil
}

func (m *mockSupplierRepo) List(_ context.Context, _, _ int) ([]domain.Supplier, int, error) {
	list := make([]domain.Supplier, 0, len(m.suppliers))
	for _, v := range m.suppliers {
		list = append(list, *v)
	}
	return list, len(list), nil
}

func (m *mockSupplierRepo) Update(_ context.Context, s *domain.Supplier) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.suppliers[s.ID]; !ok {
		return domain.ErrNotFound
	}
	m.suppliers[s.ID] = s
	return nil
}

// Stub audit service used by supplier tests (procurement_service_test.go already
// defines mockAuditSvc for that file; we define a tiny local version here).
type supplierTestAuditSvc struct{ logged int }

func (m *supplierTestAuditSvc) Log(_ context.Context, _, _ string, _, _ uuid.UUID, _ map[string]interface{}) error {
	m.logged++
	return nil
}
func (m *supplierTestAuditSvc) List(_ context.Context, _ string, _ *uuid.UUID, _, _ int) ([]domain.AuditEvent, int, error) {
	return nil, 0, nil
}
func (m *supplierTestAuditSvc) ListByEventTypes(_ context.Context, _ []string, _, _ int) ([]domain.AuditEvent, int, error) {
	return nil, 0, nil
}

func TestSupplierCreate_Success(t *testing.T) {
	repo := newMockSupplierRepo()
	audit := &supplierTestAuditSvc{}
	svc := application.NewSupplierService(repo, audit)

	sup, err := svc.Create(context.Background(), &domain.Supplier{Name: "Acme"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if sup.ID == uuid.Nil {
		t.Error("expected ID assigned")
	}
	if sup.CreatedAt.IsZero() || sup.UpdatedAt.IsZero() {
		t.Error("expected timestamps set")
	}
	if audit.logged != 1 {
		t.Errorf("expected audit log=1, got %d", audit.logged)
	}
}

func TestSupplierCreate_MissingName_Validation(t *testing.T) {
	repo := newMockSupplierRepo()
	svc := application.NewSupplierService(repo, &supplierTestAuditSvc{})

	_, err := svc.Create(context.Background(), &domain.Supplier{})
	var vErr *domain.ErrValidation
	if !errors.As(err, &vErr) || vErr.Field != "name" {
		t.Fatalf("expected validation on name, got %v", err)
	}
}

func TestSupplierCreate_RepoError(t *testing.T) {
	repo := newMockSupplierRepo()
	repo.createErr = errors.New("db down")
	svc := application.NewSupplierService(repo, &supplierTestAuditSvc{})

	_, err := svc.Create(context.Background(), &domain.Supplier{Name: "X"})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestSupplierGet_Found(t *testing.T) {
	repo := newMockSupplierRepo()
	svc := application.NewSupplierService(repo, &supplierTestAuditSvc{})

	id := uuid.New()
	repo.suppliers[id] = &domain.Supplier{ID: id, Name: "Acme"}

	sup, err := svc.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if sup.ID != id {
		t.Errorf("expected ID=%v got %v", id, sup.ID)
	}
}

func TestSupplierGet_NotFound(t *testing.T) {
	repo := newMockSupplierRepo()
	svc := application.NewSupplierService(repo, &supplierTestAuditSvc{})

	_, err := svc.Get(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSupplierList(t *testing.T) {
	repo := newMockSupplierRepo()
	svc := application.NewSupplierService(repo, &supplierTestAuditSvc{})

	for i := 0; i < 3; i++ {
		id := uuid.New()
		repo.suppliers[id] = &domain.Supplier{ID: id, Name: "Acme"}
	}

	rows, total, err := svc.List(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 3 || len(rows) != 3 {
		t.Errorf("expected 3, got total=%d rows=%d", total, len(rows))
	}
}

func TestSupplierUpdate_Success(t *testing.T) {
	repo := newMockSupplierRepo()
	audit := &supplierTestAuditSvc{}
	svc := application.NewSupplierService(repo, audit)

	id := uuid.New()
	repo.suppliers[id] = &domain.Supplier{
		ID:           id,
		Name:         "Old",
		ContactName:  "Old Contact",
		ContactEmail: "old@e.com",
		ContactPhone: "111",
		Address:      "Old Address",
		IsActive:     true,
	}

	err := svc.Update(context.Background(), &domain.Supplier{
		ID:           id,
		Name:         "New",
		ContactName:  "New Contact",
		ContactEmail: "new@e.com",
		ContactPhone: "222",
		Address:      "New Address",
		IsActive:     false,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	stored := repo.suppliers[id]
	if stored.Name != "New" || stored.ContactEmail != "new@e.com" || stored.IsActive {
		t.Errorf("fields not applied: %+v", stored)
	}
	if audit.logged != 1 {
		t.Errorf("expected audit log=1, got %d", audit.logged)
	}
}

func TestSupplierUpdate_NotFound(t *testing.T) {
	repo := newMockSupplierRepo()
	svc := application.NewSupplierService(repo, &supplierTestAuditSvc{})

	err := svc.Update(context.Background(), &domain.Supplier{ID: uuid.New(), Name: "X"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
