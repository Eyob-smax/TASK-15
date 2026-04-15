package backup_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/platform"
)

// TestVerifyIntegrity uses the same mockBackupRepo and mockAuditService declared in
// backup_service_test.go (same test package).

func TestVerifyIntegrity_Success_ReturnsRun(t *testing.T) {
	repo := newMockBackupRepo()
	audit := &mockAuditService{}
	cfg := platform.Config{BackupPath: t.TempDir(), BackupEncryptionKeyRef: "test-ref"}
	svc := newBackupService(repo, cfg, successDump, audit)

	run, err := svc.Trigger(context.Background(), nil)
	if err != nil {
		t.Fatalf("Trigger failed: %v", err)
	}

	verified, err := svc.VerifyIntegrity(context.Background(), run.ID)
	if err != nil {
		t.Fatalf("VerifyIntegrity failed: %v", err)
	}
	if verified.ID != run.ID {
		t.Errorf("expected run ID %v, got %v", run.ID, verified.ID)
	}
}

func TestVerifyIntegrity_NotFound(t *testing.T) {
	repo := newMockBackupRepo()
	audit := &mockAuditService{}
	cfg := platform.Config{BackupPath: t.TempDir()}
	svc := newBackupService(repo, cfg, successDump, audit)

	_, err := svc.VerifyIntegrity(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error when backup not found")
	}
}

func TestVerifyIntegrity_NonCompletedStatus_ReturnsError(t *testing.T) {
	repo := newMockBackupRepo()
	audit := &mockAuditService{}
	cfg := platform.Config{BackupPath: t.TempDir()}
	svc := newBackupService(repo, cfg, successDump, audit)

	run := &domain.BackupRun{
		ID:     uuid.New(),
		Status: domain.BackupStatusRunning,
	}
	repo.runs[run.ID] = run

	_, err := svc.VerifyIntegrity(context.Background(), run.ID)
	if err == nil {
		t.Fatal("expected error for non-completed status")
	}
}

func TestVerifyIntegrity_EmptyChecksum_ReturnsError(t *testing.T) {
	repo := newMockBackupRepo()
	audit := &mockAuditService{}
	cfg := platform.Config{BackupPath: t.TempDir()}
	svc := newBackupService(repo, cfg, successDump, audit)

	run := &domain.BackupRun{
		ID:       uuid.New(),
		Status:   domain.BackupStatusCompleted,
		Checksum: "",
	}
	repo.runs[run.ID] = run

	_, err := svc.VerifyIntegrity(context.Background(), run.ID)
	if err == nil {
		t.Fatal("expected error when Checksum empty")
	}
}

func TestVerifyIntegrity_MissingFile_ReturnsError(t *testing.T) {
	repo := newMockBackupRepo()
	audit := &mockAuditService{}
	cfg := platform.Config{BackupPath: t.TempDir()}
	svc := newBackupService(repo, cfg, successDump, audit)

	run := &domain.BackupRun{
		ID:          uuid.New(),
		Status:      domain.BackupStatusCompleted,
		Checksum:    "abcd",
		ArchivePath: filepath.Join(t.TempDir(), "missing.dump"),
	}
	repo.runs[run.ID] = run

	_, err := svc.VerifyIntegrity(context.Background(), run.ID)
	if err == nil {
		t.Fatal("expected error when archive file missing")
	}
}

func TestVerifyIntegrity_ChecksumMismatch_ReturnsRunAndError(t *testing.T) {
	repo := newMockBackupRepo()
	audit := &mockAuditService{}
	cfg := platform.Config{BackupPath: t.TempDir(), BackupEncryptionKeyRef: "test-ref"}
	svc := newBackupService(repo, cfg, successDump, audit)

	run, err := svc.Trigger(context.Background(), nil)
	if err != nil {
		t.Fatalf("Trigger failed: %v", err)
	}
	// Tamper the file content so checksum changes.
	if err := os.WriteFile(run.ArchivePath, []byte("tampered bytes"), 0600); err != nil {
		t.Fatalf("tamper write failed: %v", err)
	}

	verified, err := svc.VerifyIntegrity(context.Background(), run.ID)
	if err == nil {
		t.Fatal("expected checksum mismatch error")
	}
	if verified == nil {
		t.Error("expected non-nil run even on mismatch")
	}
}
