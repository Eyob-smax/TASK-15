package jobs

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
)

// BackupJob runs database backups on a 24-hour interval.
type BackupJob struct {
	svc      application.BackupService
	logger   *slog.Logger
	interval time.Duration
}

func NewBackupJob(svc application.BackupService, logger *slog.Logger) *BackupJob {
	return &BackupJob{svc: svc, logger: logger, interval: 24 * time.Hour}
}

func NewBackupJobWithInterval(svc application.BackupService, logger *slog.Logger, interval time.Duration) *BackupJob {
	return &BackupJob{svc: svc, logger: logger, interval: interval}
}

func (j *BackupJob) Run(ctx context.Context) {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run, err := j.svc.Trigger(ctx, nil) // nil = system actor
			if err != nil {
				j.logger.Error("backup job error", "error", err)
				continue
			}
			j.logger.Info("backup completed", "run_id", run.ID, "status", string(run.Status))
		}
	}
}

// VarianceDeadlineJob scans open variances past their due date every hour.
type VarianceDeadlineJob struct {
	svc      application.VarianceService
	logger   *slog.Logger
	interval time.Duration
}

func NewVarianceDeadlineJob(svc application.VarianceService, logger *slog.Logger) *VarianceDeadlineJob {
	return &VarianceDeadlineJob{svc: svc, logger: logger, interval: time.Hour}
}

func NewVarianceDeadlineJobWithInterval(svc application.VarianceService, logger *slog.Logger, interval time.Duration) *VarianceDeadlineJob {
	return &VarianceDeadlineJob{svc: svc, logger: logger, interval: interval}
}

func (j *VarianceDeadlineJob) Run(ctx context.Context) {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count, err := j.svc.EscalateOverdue(ctx)
			if err != nil {
				j.logger.Error("variance escalation error", "error", err)
				continue
			}
			if count > 0 {
				j.logger.Warn("escalated overdue variances", "count", count)
			}
		}
	}
}

// RetentionCleanupJob enforces retention policies every 24 hours by hard-deleting
// records older than their configured retention window.
type RetentionCleanupJob struct {
	svc      application.RetentionService
	logger   *slog.Logger
	interval time.Duration
}

func NewRetentionCleanupJob(svc application.RetentionService, logger *slog.Logger) *RetentionCleanupJob {
	return &RetentionCleanupJob{svc: svc, logger: logger, interval: 24 * time.Hour}
}

func NewRetentionCleanupJobWithInterval(svc application.RetentionService, logger *slog.Logger, interval time.Duration) *RetentionCleanupJob {
	return &RetentionCleanupJob{svc: svc, logger: logger, interval: interval}
}

func (j *RetentionCleanupJob) Run(ctx context.Context) {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := j.svc.RunCleanup(ctx); err != nil {
				j.logger.Error("retention cleanup error", "error", err)
				continue
			}
			j.logger.Info("retention cleanup scan completed")
		}
	}
}

// BiometricKeyRotationJob ensures biometric encryption keys are present and
// rotated on schedule when the biometric module is enabled.
type BiometricKeyRotationJob struct {
	svc          application.BiometricService
	logger       *slog.Logger
	rotationDays int
	interval     time.Duration
}

func NewBiometricKeyRotationJob(svc application.BiometricService, logger *slog.Logger, rotationDays int) *BiometricKeyRotationJob {
	return &BiometricKeyRotationJob{
		svc:          svc,
		logger:       logger,
		rotationDays: rotationDays,
		interval:     24 * time.Hour,
	}
}

func NewBiometricKeyRotationJobWithInterval(svc application.BiometricService, logger *slog.Logger, rotationDays int, interval time.Duration) *BiometricKeyRotationJob {
	return &BiometricKeyRotationJob{
		svc:          svc,
		logger:       logger,
		rotationDays: rotationDays,
		interval:     interval,
	}
}

func (j *BiometricKeyRotationJob) Run(ctx context.Context) {
	if err := j.ensureKeyState(ctx); err != nil {
		j.logger.Error("biometric key rotation check failed", "error", err)
	}

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := j.ensureKeyState(ctx); err != nil {
				j.logger.Error("biometric key rotation check failed", "error", err)
			}
		}
	}
}

func (j *BiometricKeyRotationJob) ensureKeyState(ctx context.Context) error {
	key, err := j.svc.GetActiveKey(ctx)
	if errors.Is(err, domain.ErrNotFound) {
		rotated, rotateErr := j.svc.RotateKey(ctx, domain.SystemActorID)
		if rotateErr != nil {
			return rotateErr
		}
		j.logger.Info("bootstrapped biometric encryption key", "key_id", rotated.ID)
		return nil
	}
	if err != nil {
		return err
	}

	if key.NeedsRotation(j.rotationDays) || time.Now().UTC().After(key.ExpiresAt) {
		rotated, rotateErr := j.svc.RotateKey(ctx, domain.SystemActorID)
		if rotateErr != nil {
			return rotateErr
		}
		j.logger.Info("rotated biometric encryption key", "key_id", rotated.ID)
	}

	return nil
}
