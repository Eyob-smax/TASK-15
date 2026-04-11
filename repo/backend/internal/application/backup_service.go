package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/platform"
	"fitcommerce/internal/security"
	"fitcommerce/internal/store"
)

// DumpFunc is a function that performs a database dump from connStr to destPath.
type DumpFunc func(ctx context.Context, connStr, destPath string) error

// BackupServiceImpl implements BackupService.
type BackupServiceImpl struct {
	backupRepo store.BackupRepository
	cfg        platform.Config
	dump       DumpFunc
	audit      AuditService
}

// NewBackupService creates a BackupServiceImpl with the given dependencies.
func NewBackupService(
	backupRepo store.BackupRepository,
	cfg platform.Config,
	dump DumpFunc,
	auditSvc AuditService,
) *BackupServiceImpl {
	return &BackupServiceImpl{
		backupRepo: backupRepo,
		cfg:        cfg,
		dump:       dump,
		audit:      auditSvc,
	}
}

// Trigger initiates a new database backup run. performedBy is nil for scheduled/system
// runs and non-nil for manually triggered runs. The actor is recorded only in the audit
// event; it is not stored on the BackupRun record itself.
func (s *BackupServiceImpl) Trigger(ctx context.Context, performedBy *uuid.UUID) (*domain.BackupRun, error) {
	run := &domain.BackupRun{
		ID:        uuid.New(),
		Status:    domain.BackupStatusRunning,
		StartedAt: time.Now().UTC(),
	}

	if err := s.backupRepo.Create(ctx, run); err != nil {
		return nil, err
	}

	destPath := filepath.Join(
		s.cfg.BackupPath,
		fmt.Sprintf("backup_%s.dump", run.StartedAt.Format("20060102_150405")),
	)

	if err := os.MkdirAll(s.cfg.BackupPath, 0755); err != nil {
		run.Status = domain.BackupStatusFailed
		_ = s.backupRepo.Update(ctx, run) // best-effort; ignore error
		return run, fmt.Errorf("backup_service.Trigger mkdir backup path: %w", err)
	}

	if err := s.dump(ctx, s.cfg.DatabaseURL, destPath); err != nil {
		run.Status = domain.BackupStatusFailed
		_ = s.backupRepo.Update(ctx, run) // best-effort; ignore error
		return run, fmt.Errorf("backup_service.Trigger dump: %w", err)
	}

	// Encryption is required. Fail the backup if no key reference is configured so
	// unencrypted archives are never silently stored.
	if s.cfg.BackupEncryptionKeyRef == "" {
		run.Status = domain.BackupStatusFailed
		_ = s.backupRepo.Update(ctx, run)
		return run, fmt.Errorf("backup_service.Trigger: FC_BACKUP_ENCRYPTION_KEY_REF is not set; backup encryption is mandatory")
	}
	key, err := security.DeriveKeyFromRef(s.cfg.BackupEncryptionKeyRef)
	if err != nil {
		run.Status = domain.BackupStatusFailed
		_ = s.backupRepo.Update(ctx, run)
		return run, err
	}
	if err := encryptFile(destPath, key); err != nil {
		run.Status = domain.BackupStatusFailed
		_ = s.backupRepo.Update(ctx, run)
		return run, fmt.Errorf("backup_service.Trigger encrypt: %w", err)
	}
	run.EncryptionKeyRef = s.cfg.BackupEncryptionKeyRef

	// Persist metadata for the final encrypted archive rather than the plaintext dump.
	if stat, err := os.Stat(destPath); err == nil {
		run.FileSize = stat.Size()
	}
	checksum, err := sha256HexFile(destPath)
	if err != nil {
		run.Status = domain.BackupStatusFailed
		_ = s.backupRepo.Update(ctx, run)
		return run, fmt.Errorf("backup_service.Trigger checksum: %w", err)
	}
	run.Checksum = checksum
	run.ChecksumAlgorithm = "sha256"

	now := time.Now().UTC()
	run.Status = domain.BackupStatusCompleted
	run.ArchivePath = destPath
	run.CompletedAt = &now

	if err := s.backupRepo.Update(ctx, run); err != nil {
		return run, fmt.Errorf("backup_service.Trigger update: %w", err)
	}

	actorID := domain.SystemActorID
	if performedBy != nil {
		actorID = *performedBy
	}

	if err := s.audit.Log(ctx, "backup.completed", "backup_run", run.ID, actorID, map[string]interface{}{
		"path":      destPath,
		"checksum":  checksum,
		"file_size": run.FileSize,
	}); err != nil {
		slog.Default().Warn("audit log failed", "event", "backup.completed", "error", err)
	}

	return run, nil
}

// VerifyIntegrity re-computes the SHA-256 checksum of the backup archive and
// compares it against the stored value. Returns the BackupRun on success or an
// error if the file is missing, the backup is not in a completed state, or the
// checksum does not match.
func (s *BackupServiceImpl) VerifyIntegrity(ctx context.Context, id uuid.UUID) (*domain.BackupRun, error) {
	run, err := s.backupRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if run.Status != domain.BackupStatusCompleted || run.Checksum == "" {
		return nil, fmt.Errorf("backup_service.VerifyIntegrity: backup %s is not in a verifiable state (status=%s)", id, run.Status)
	}
	actual, err := sha256HexFile(run.ArchivePath)
	if err != nil {
		return nil, fmt.Errorf("backup_service.VerifyIntegrity: %w", err)
	}
	if actual != run.Checksum {
		return run, fmt.Errorf("backup_service.VerifyIntegrity: checksum mismatch — stored=%s actual=%s", run.Checksum, actual)
	}
	return run, nil
}

// GetByID retrieves a backup run record by ID.
func (s *BackupServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.BackupRun, error) {
	return s.backupRepo.GetByID(ctx, id)
}

// List returns a paginated list of backup run records.
func (s *BackupServiceImpl) List(ctx context.Context, page, pageSize int) ([]domain.BackupRun, int, error) {
	return s.backupRepo.List(ctx, page, pageSize)
}

// sha256HexFile streams the file at path through SHA-256 and returns the hex-encoded digest.
func sha256HexFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// encryptFile reads the file at path, encrypts it with AES-GCM using key, and
// overwrites the file with the ciphertext.
func encryptFile(path string, key []byte) error {
	plaintext, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	ciphertext, err := security.EncryptAESGCM(plaintext, key)
	if err != nil {
		return err
	}
	return os.WriteFile(path, ciphertext, 0600)
}

// Compile-time interface assertion.
var _ BackupService = (*BackupServiceImpl)(nil)
