package admin_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
)

// --- Mock AuditRepository ---

type mockAuditRepo struct {
	events         []domain.AuditEvent
	latestHash     string
	latestHashErr  error
	createErr      error
	listErr        error
	listByTypesErr error
}

func (m *mockAuditRepo) Create(_ context.Context, event *domain.AuditEvent) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.events = append(m.events, *event)
	return nil
}
func (m *mockAuditRepo) List(_ context.Context, entityType string, entityID *uuid.UUID, _, _ int) ([]domain.AuditEvent, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var out []domain.AuditEvent
	for _, e := range m.events {
		if entityType != "" && e.EntityType != entityType {
			continue
		}
		if entityID != nil && e.EntityID != *entityID {
			continue
		}
		out = append(out, e)
	}
	return out, len(out), nil
}
func (m *mockAuditRepo) ListByEventTypes(_ context.Context, types []string, _, _ int) ([]domain.AuditEvent, int, error) {
	if m.listByTypesErr != nil {
		return nil, 0, m.listByTypesErr
	}
	set := map[string]bool{}
	for _, t := range types {
		set[t] = true
	}
	var out []domain.AuditEvent
	for _, e := range m.events {
		if set[e.EventType] {
			out = append(out, e)
		}
	}
	return out, len(out), nil
}
func (m *mockAuditRepo) GetLatestHash(_ context.Context) (string, error) {
	if m.latestHashErr != nil {
		return "", m.latestHashErr
	}
	return m.latestHash, nil
}

func TestAuditService_Log_Success(t *testing.T) {
	repo := &mockAuditRepo{}
	svc := application.NewAuditService(repo)

	err := svc.Log(context.Background(), "user.created", "user", uuid.New(), uuid.New(), map[string]interface{}{"email": "x@e.com"})
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}
	if len(repo.events) != 1 {
		t.Fatalf("expected 1 event persisted, got %d", len(repo.events))
	}
	if repo.events[0].EventType != "user.created" {
		t.Errorf("unexpected event type: %q", repo.events[0].EventType)
	}
	if repo.events[0].IntegrityHash == "" {
		t.Error("expected IntegrityHash to be computed")
	}
}

func TestAuditService_Log_HashError(t *testing.T) {
	repo := &mockAuditRepo{latestHashErr: errors.New("db broken")}
	svc := application.NewAuditService(repo)

	err := svc.Log(context.Background(), "x", "y", uuid.New(), uuid.New(), nil)
	if err == nil {
		t.Fatal("expected error when GetLatestHash fails")
	}
}

func TestAuditService_Log_CreateError(t *testing.T) {
	repo := &mockAuditRepo{createErr: errors.New("db broken")}
	svc := application.NewAuditService(repo)

	err := svc.Log(context.Background(), "x", "y", uuid.New(), uuid.New(), nil)
	if err == nil {
		t.Fatal("expected error when Create fails")
	}
}

func TestAuditService_List_FiltersByEntity(t *testing.T) {
	entityID := uuid.New()
	repo := &mockAuditRepo{events: []domain.AuditEvent{
		{ID: uuid.New(), EventType: "x", EntityType: "user", EntityID: entityID},
		{ID: uuid.New(), EventType: "y", EntityType: "user", EntityID: uuid.New()},
		{ID: uuid.New(), EventType: "z", EntityType: "order", EntityID: entityID},
	}}
	svc := application.NewAuditService(repo)

	rows, total, err := svc.List(context.Background(), "user", &entityID, 1, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 || len(rows) != 1 {
		t.Errorf("expected 1 row after filter, got total=%d rows=%d", total, len(rows))
	}
}

func TestAuditService_List_RepoError(t *testing.T) {
	repo := &mockAuditRepo{listErr: errors.New("db broken")}
	svc := application.NewAuditService(repo)

	_, _, err := svc.List(context.Background(), "user", nil, 1, 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAuditService_ListByEventTypes_FiltersByTypes(t *testing.T) {
	repo := &mockAuditRepo{events: []domain.AuditEvent{
		{ID: uuid.New(), EventType: "login.success"},
		{ID: uuid.New(), EventType: "login.failure"},
		{ID: uuid.New(), EventType: "order.created"},
	}}
	svc := application.NewAuditService(repo)

	rows, total, err := svc.ListByEventTypes(context.Background(), []string{"login.success", "login.failure"}, 1, 10)
	if err != nil {
		t.Fatalf("ListByEventTypes failed: %v", err)
	}
	if total != 2 || len(rows) != 2 {
		t.Errorf("expected 2 rows, got total=%d rows=%d", total, len(rows))
	}
}

func TestAuditService_ListByEventTypes_RepoError(t *testing.T) {
	repo := &mockAuditRepo{listByTypesErr: errors.New("db broken")}
	svc := application.NewAuditService(repo)
	_, _, err := svc.ListByEventTypes(context.Background(), []string{"x"}, 1, 10)
	if err == nil {
		t.Fatal("expected error")
	}
}
