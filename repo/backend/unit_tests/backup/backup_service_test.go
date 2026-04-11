package backup_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
	"fitcommerce/internal/platform"
)

// --- Mock repositories ---

type mockBackupRepo struct {
	runs       map[uuid.UUID]*domain.BackupRun
	updateCalls int
}

func newMockBackupRepo() *mockBackupRepo {
	return &mockBackupRepo{runs: make(map[uuid.UUID]*domain.BackupRun)}
}

func (m *mockBackupRepo) Create(_ context.Context, run *domain.BackupRun) error {
	m.runs[run.ID] = run
	return nil
}

func (m *mockBackupRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.BackupRun, error) {
	run, ok := m.runs[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return run, nil
}

func (m *mockBackupRepo) List(_ context.Context, _, _ int) ([]domain.BackupRun, int, error) {
	list := make([]domain.BackupRun, 0, len(m.runs))
	for _, v := range m.runs {
		list = append(list, *v)
	}
	return list, len(list), nil
}

func (m *mockBackupRepo) Update(_ context.Context, run *domain.BackupRun) error {
	m.updateCalls++
	if _, ok := m.runs[run.ID]; !ok {
		return domain.ErrNotFound
	}
	m.runs[run.ID] = run
	return nil
}

// --- Mock audit service ---

type mockAuditService struct {
	logCalled int
}

func (m *mockAuditService) Log(_ context.Context, _, _ string, _, _ uuid.UUID, _ map[string]interface{}) error {
	m.logCalled++
	return nil
}

func (m *mockAuditService) List(_ context.Context, _ string, _ *uuid.UUID, _, _ int) ([]domain.AuditEvent, int, error) {
	return nil, 0, nil
}

func (m *mockAuditService) ListByEventTypes(_ context.Context, _ []string, _, _ int) ([]domain.AuditEvent, int, error) {
	return nil, 0, nil
}

// --- DumpFunc helpers ---

func successDump(_ context.Context, _, destPath string) error {
	return os.WriteFile(destPath, []byte("backup content"), 0600)
}

func failDump(_ context.Context, _, _ string) error {
	return errors.New("pg_dump: connection refused")
}

// --- Helper to build BackupService ---

func newBackupService(repo *mockBackupRepo, cfg platform.Config, dump application.DumpFunc, audit *mockAuditService) application.BackupService {
	return application.NewBackupService(repo, cfg, dump, audit)
}

// --- Tests ---

func TestTrigger_NilActor_Success(t *testing.T) {
	repo := newMockBackupRepo()
	audit := &mockAuditService{}
	cfg := platform.Config{BackupPath: t.TempDir(), BackupEncryptionKeyRef: "test-ref"}

	svc := newBackupService(repo, cfg, successDump, audit)

	run, err := svc.Trigger(context.Background(), nil)
	if err != nil {
		t.Fatalf("Trigger failed: %v", err)
	}
	if run.Status != domain.BackupStatusCompleted {
		t.Errorf("expected status=completed, got %v", run.Status)
	}
	if run.Checksum == "" {
		t.Error("expected non-empty Checksum after successful backup")
	}
	if audit.logCalled == 0 {
		t.Error("expected auditSvc.Log to be called")
	}
}

func TestTrigger_WithUserActor(t *testing.T) {
	repo := newMockBackupRepo()
	audit := &mockAuditService{}
	cfg := platform.Config{BackupPath: t.TempDir(), BackupEncryptionKeyRef: "test-ref"}

	svc := newBackupService(repo, cfg, successDump, audit)

	actorID := uuid.New()
	run, err := svc.Trigger(context.Background(), &actorID)
	if err != nil {
		t.Fatalf("Trigger failed: %v", err)
	}
	if run.Status != domain.BackupStatusCompleted {
		t.Errorf("expected status=completed, got %v", run.Status)
	}
}

func TestTrigger_DumpFail_StatusFailed(t *testing.T) {
	repo := newMockBackupRepo()
	audit := &mockAuditService{}
	cfg := platform.Config{BackupPath: t.TempDir()}

	svc := newBackupService(repo, cfg, failDump, audit)

	run, err := svc.Trigger(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error from failed dump, got nil")
	}
	if run == nil {
		t.Fatal("expected non-nil BackupRun even on failure")
	}
	if run.Status != domain.BackupStatusFailed {
		t.Errorf("expected status=failed, got %v", run.Status)
	}
	// backupRepo.Update should have been called with the failed run.
	if repo.updateCalls == 0 {
		t.Error("expected backupRepo.Update to be called after dump failure")
	}
}

func TestTrigger_WithEncryptionKeyRef(t *testing.T) {
	repo := newMockBackupRepo()
	audit := &mockAuditService{}
	cfg := platform.Config{
		BackupPath:             t.TempDir(),
		BackupEncryptionKeyRef: "test-ref",
	}

	svc := newBackupService(repo, cfg, successDump, audit)

	run, err := svc.Trigger(context.Background(), nil)
	if err != nil {
		t.Fatalf("Trigger with encryption key ref failed: %v", err)
	}
	if run.EncryptionKeyRef != "test-ref" {
		t.Errorf("expected EncryptionKeyRef=%q, got %q", "test-ref", run.EncryptionKeyRef)
	}
}

func TestList_DelegatesToRepo(t *testing.T) {
	repo := newMockBackupRepo()
	audit := &mockAuditService{}
	cfg := platform.Config{BackupPath: t.TempDir(), BackupEncryptionKeyRef: "test-ref"}

	svc := newBackupService(repo, cfg, successDump, audit)

	// Trigger a run so there is something to list.
	_, err := svc.Trigger(context.Background(), nil)
	if err != nil {
		t.Fatalf("Trigger failed: %v", err)
	}

	runs, total, err := svc.List(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total == 0 {
		t.Error("expected at least 1 backup run in List result")
	}
	_ = runs
}

func TestGetByID_DelegatesToRepo(t *testing.T) {
	repo := newMockBackupRepo()
	audit := &mockAuditService{}
	cfg := platform.Config{BackupPath: t.TempDir(), BackupEncryptionKeyRef: "test-ref"}

	svc := newBackupService(repo, cfg, successDump, audit)

	run, err := svc.Trigger(context.Background(), nil)
	if err != nil {
		t.Fatalf("Trigger failed: %v", err)
	}

	fetched, err := svc.GetByID(context.Background(), run.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if fetched.ID != run.ID {
		t.Errorf("expected ID=%v, got %v", run.ID, fetched.ID)
	}
}

func TestTrigger_WithoutEncryptionKeyRef_StatusFailed(t *testing.T) {
	repo := newMockBackupRepo()
	audit := &mockAuditService{}
	cfg := platform.Config{BackupPath: t.TempDir()}

	svc := newBackupService(repo, cfg, successDump, audit)

	run, err := svc.Trigger(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error when backup encryption key ref is missing, got nil")
	}
	if run == nil {
		t.Fatal("expected non-nil BackupRun when encryption config is invalid")
	}
	if run.Status != domain.BackupStatusFailed {
		t.Errorf("expected status=failed, got %v", run.Status)
	}
	if repo.updateCalls == 0 {
		t.Error("expected backupRepo.Update to be called after encryption configuration failure")
	}
}
