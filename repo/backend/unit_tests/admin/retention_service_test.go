package admin_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
)

// --- Mock RetentionRepository ---

type mockRetentionRepo struct {
	policies     map[string]*domain.RetentionPolicy
	createErr    error
	updateErr    error
	listErr      error
	getErr       error
	deleteByIDErr error
}

func newMockRetentionRepo() *mockRetentionRepo {
	return &mockRetentionRepo{policies: make(map[string]*domain.RetentionPolicy)}
}

func (m *mockRetentionRepo) Create(_ context.Context, p *domain.RetentionPolicy) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.policies[p.EntityType] = p
	return nil
}
func (m *mockRetentionRepo) GetByEntityType(_ context.Context, entityType string) (*domain.RetentionPolicy, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	p, ok := m.policies[entityType]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return p, nil
}
func (m *mockRetentionRepo) List(_ context.Context) ([]domain.RetentionPolicy, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	out := make([]domain.RetentionPolicy, 0, len(m.policies))
	for _, p := range m.policies {
		out = append(out, *p)
	}
	return out, nil
}
func (m *mockRetentionRepo) Update(_ context.Context, p *domain.RetentionPolicy) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.policies[p.EntityType] = p
	return nil
}
func (m *mockRetentionRepo) DeleteByIDs(_ context.Context, _ string, _ []uuid.UUID) (int64, error) {
	if m.deleteByIDErr != nil {
		return 0, m.deleteByIDErr
	}
	return 0, nil
}

// Minimal audit stub for retention tests.
type retentionAuditStub struct {
	events []string
	err    error
}

func (m *retentionAuditStub) Log(_ context.Context, event, _ string, _, _ uuid.UUID, _ map[string]interface{}) error {
	if m.err != nil {
		return m.err
	}
	m.events = append(m.events, event)
	return nil
}
func (m *retentionAuditStub) List(_ context.Context, _ string, _ *uuid.UUID, _, _ int) ([]domain.AuditEvent, int, error) {
	return nil, 0, nil
}
func (m *retentionAuditStub) ListByEventTypes(_ context.Context, _ []string, _, _ int) ([]domain.AuditEvent, int, error) {
	return nil, 0, nil
}

func TestRetentionService_GetByEntityType_Found(t *testing.T) {
	repo := newMockRetentionRepo()
	repo.policies["sessions"] = &domain.RetentionPolicy{EntityType: "sessions", RetentionDays: 30}
	svc := application.NewRetentionService(repo, &retentionAuditStub{}, nil)

	got, err := svc.GetByEntityType(context.Background(), "sessions")
	if err != nil {
		t.Fatalf("GetByEntityType failed: %v", err)
	}
	if got.RetentionDays != 30 {
		t.Errorf("expected 30, got %d", got.RetentionDays)
	}
}

func TestRetentionService_GetByEntityType_NotFound(t *testing.T) {
	repo := newMockRetentionRepo()
	svc := application.NewRetentionService(repo, &retentionAuditStub{}, nil)

	_, err := svc.GetByEntityType(context.Background(), "unknown")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRetentionService_List_ReturnsAll(t *testing.T) {
	repo := newMockRetentionRepo()
	repo.policies["sessions"] = &domain.RetentionPolicy{EntityType: "sessions", RetentionDays: 30}
	repo.policies["captcha"] = &domain.RetentionPolicy{EntityType: "captcha", RetentionDays: 7}
	svc := application.NewRetentionService(repo, &retentionAuditStub{}, nil)

	rows, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}
}

func TestRetentionService_List_RepoError(t *testing.T) {
	repo := newMockRetentionRepo()
	repo.listErr = errors.New("db down")
	svc := application.NewRetentionService(repo, &retentionAuditStub{}, nil)

	_, err := svc.List(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRetentionService_Update_CreatesWhenNotExists(t *testing.T) {
	repo := newMockRetentionRepo()
	audit := &retentionAuditStub{}
	svc := application.NewRetentionService(repo, audit, nil)

	p := &domain.RetentionPolicy{EntityType: "sessions", RetentionDays: 60, Description: "30-day"}
	if err := svc.Update(context.Background(), p); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	stored := repo.policies["sessions"]
	if stored == nil {
		t.Fatal("expected policy stored")
	}
	if stored.ID == uuid.Nil {
		t.Error("expected ID assigned on create")
	}
	if stored.RetentionDays != 60 {
		t.Errorf("expected 60, got %d", stored.RetentionDays)
	}
	if len(audit.events) != 1 || audit.events[0] != "retention.policy.updated" {
		t.Errorf("expected audit event, got %v", audit.events)
	}
}

func TestRetentionService_Update_UpdatesWhenExists(t *testing.T) {
	repo := newMockRetentionRepo()
	audit := &retentionAuditStub{}
	existing := &domain.RetentionPolicy{
		ID:            uuid.New(),
		EntityType:    "sessions",
		RetentionDays: 30,
		Description:   "original",
	}
	repo.policies["sessions"] = existing

	svc := application.NewRetentionService(repo, audit, nil)
	err := svc.Update(context.Background(), &domain.RetentionPolicy{
		EntityType:    "sessions",
		RetentionDays: 90,
		Description:   "updated",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	stored := repo.policies["sessions"]
	if stored.RetentionDays != 90 {
		t.Errorf("expected 90, got %d", stored.RetentionDays)
	}
	if stored.Description != "updated" {
		t.Errorf("description not applied: %q", stored.Description)
	}
	if stored.ID != existing.ID {
		t.Error("ID should stay the same for upsert-update path")
	}
}

func TestRetentionService_Update_GetError(t *testing.T) {
	repo := newMockRetentionRepo()
	repo.getErr = errors.New("db broken")
	svc := application.NewRetentionService(repo, &retentionAuditStub{}, nil)

	err := svc.Update(context.Background(), &domain.RetentionPolicy{EntityType: "sessions", RetentionDays: 10})
	if err == nil {
		t.Fatal("expected error from Get")
	}
}

func TestRetentionService_Update_CreateError(t *testing.T) {
	repo := newMockRetentionRepo()
	repo.createErr = errors.New("create failed")
	svc := application.NewRetentionService(repo, &retentionAuditStub{}, nil)

	err := svc.Update(context.Background(), &domain.RetentionPolicy{EntityType: "sessions", RetentionDays: 10})
	if err == nil {
		t.Fatal("expected error from Create")
	}
}

func TestRetentionService_Update_UpdateError(t *testing.T) {
	repo := newMockRetentionRepo()
	repo.policies["sessions"] = &domain.RetentionPolicy{
		ID:            uuid.New(),
		EntityType:    "sessions",
		RetentionDays: 30,
	}
	repo.updateErr = errors.New("update failed")
	svc := application.NewRetentionService(repo, &retentionAuditStub{}, nil)

	err := svc.Update(context.Background(), &domain.RetentionPolicy{EntityType: "sessions", RetentionDays: 90})
	if err == nil {
		t.Fatal("expected error from Update")
	}
}
