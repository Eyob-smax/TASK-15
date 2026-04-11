package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/platform"
	"fitcommerce/internal/security"
	"fitcommerce/internal/store"
)

// BiometricServiceImpl implements BiometricService.
type BiometricServiceImpl struct {
	biometricRepo store.BiometricRepository
	encKeyRepo    store.EncryptionKeyRepository
	auditSvc      AuditService
	txPool        *pgxpool.Pool
	rotationDays  int
	cfg           platform.Config
}

// NewBiometricService creates a BiometricServiceImpl backed by the given repositories.
func NewBiometricService(
	biometricRepo store.BiometricRepository,
	encKeyRepo store.EncryptionKeyRepository,
	auditSvc AuditService,
	txPool *pgxpool.Pool,
	rotationDays int,
	cfg platform.Config,
) *BiometricServiceImpl {
	if rotationDays <= 0 {
		rotationDays = 90
	}
	return &BiometricServiceImpl{
		biometricRepo: biometricRepo,
		encKeyRepo:    encKeyRepo,
		auditSvc:      auditSvc,
		txPool:        txPool,
		rotationDays:  rotationDays,
		cfg:           cfg,
	}
}

// unwrapDEK derives the KEK from the configured master key reference and uses it
// to unwrap the DEK stored in the given EncryptionKey.
func (s *BiometricServiceImpl) unwrapDEK(key *domain.EncryptionKey) ([]byte, error) {
	if len(key.WrappedDEK) == 0 {
		return nil, fmt.Errorf("biometric: encryption key %s has no wrapped DEK — key rotation required", key.ID)
	}
	kek, err := security.DeriveKeyFromRef(s.cfg.BiometricMasterKeyRef)
	if err != nil {
		return nil, fmt.Errorf("biometric: KEK derivation failed: %w", err)
	}
	dek, err := security.UnwrapDEK(key.WrappedDEK, kek)
	if err != nil {
		return nil, fmt.Errorf("biometric: DEK unwrap failed: %w", err)
	}
	return dek, nil
}

// Register creates a new biometric enrollment for the given user. The
// templateRef is an opaque reference; raw biometric data never enters the
// application layer. The reference is encrypted with AES-256-GCM using the
// active biometric encryption key before storage.
func (s *BiometricServiceImpl) Register(ctx context.Context, userID uuid.UUID, templateRef string) (*domain.BiometricEnrollment, error) {
	activeKey, err := s.ensureActiveKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("biometric.Register: no active encryption key: %w", err)
	}

	aesKey, err := s.unwrapDEK(activeKey)
	if err != nil {
		return nil, fmt.Errorf("biometric.Register: %w", err)
	}

	encrypted, err := security.EncryptAESGCM([]byte(templateRef), aesKey)
	if err != nil {
		return nil, fmt.Errorf("biometric.Register: encryption failed: %w", err)
	}

	now := time.Now().UTC()
	enrollment := &domain.BiometricEnrollment{
		ID:              uuid.New(),
		UserID:          userID,
		EncryptedData:   encrypted,
		EncryptionKeyID: activeKey.ID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.biometricRepo.Create(ctx, enrollment); err != nil {
		return nil, err
	}

	_ = s.auditSvc.Log(ctx, "biometric.registered", "biometric_enrollment", enrollment.ID, userID, map[string]interface{}{})

	return enrollment, nil
}

// GetByUser retrieves the biometric enrollment for the given user, with the
// encrypted data replaced by a redaction placeholder so raw data is never
// exposed outside this layer.
func (s *BiometricServiceImpl) GetByUser(ctx context.Context, userID uuid.UUID) (*domain.BiometricEnrollment, error) {
	enrollment, err := s.biometricRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	copy := *enrollment
	copy.EncryptedData = []byte(security.RedactBiometric())
	return &copy, nil
}

// Revoke clears the biometric data for the given user's enrollment record.
func (s *BiometricServiceImpl) Revoke(ctx context.Context, userID uuid.UUID) error {
	enrollment, err := s.biometricRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	enrollment.EncryptedData = []byte{}
	enrollment.UpdatedAt = time.Now().UTC()

	if err := s.biometricRepo.Update(ctx, enrollment); err != nil {
		return err
	}

	_ = s.auditSvc.Log(ctx, "biometric.revoked", "biometric_enrollment", enrollment.ID, userID, map[string]interface{}{})

	return nil
}

// RotateKey rotates the active biometric envelope key and re-encrypts existing
// enrollment references under the new active key inside one transaction.
func (s *BiometricServiceImpl) RotateKey(ctx context.Context, performedBy uuid.UUID) (*domain.EncryptionKey, error) {
	var newKey *domain.EncryptionKey

	err := withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		oldKey, err := s.encKeyRepo.GetActive(txCtx, "biometric")
		if err != nil && err != domain.ErrNotFound {
			return err
		}

		// Generate a fresh random DEK for the new key slot.
		newDEK, err := security.GenerateAESKey()
		if err != nil {
			return fmt.Errorf("biometric.RotateKey: DEK generation failed: %w", err)
		}
		kek, err := security.DeriveKeyFromRef(s.cfg.BiometricMasterKeyRef)
		if err != nil {
			return fmt.Errorf("biometric.RotateKey: KEK derivation failed: %w", err)
		}
		wrappedDEK, err := security.WrapDEK(newDEK, kek)
		if err != nil {
			return fmt.Errorf("biometric.RotateKey: DEK wrap failed: %w", err)
		}

		newKey = &domain.EncryptionKey{
			ID:           uuid.New(),
			KeyReference: uuid.New().String(),
			WrappedDEK:   wrappedDEK,
			Purpose:      "biometric",
			Status:       domain.EncryptionKeyStatusActive,
			ActivatedAt:  time.Now().UTC(),
			ExpiresAt:    time.Now().UTC().AddDate(0, 0, s.rotationDays),
		}
		newAESKey := newDEK

		var oldAESKey []byte
		if oldKey != nil && err != domain.ErrNotFound {
			oldAESKey, err = s.unwrapDEK(oldKey)
			if err != nil {
				return fmt.Errorf("biometric.RotateKey old key unwrap failed: %w", err)
			}
		}

		if err := s.encKeyRepo.Create(txCtx, newKey); err != nil {
			return err
		}

		enrollments, err := s.biometricRepo.List(txCtx)
		if err != nil {
			return err
		}
		for i := range enrollments {
			if len(enrollments[i].EncryptedData) == 0 || oldAESKey == nil {
				enrollments[i].EncryptionKeyID = newKey.ID
				enrollments[i].UpdatedAt = time.Now().UTC()
				if err := s.biometricRepo.Update(txCtx, &enrollments[i]); err != nil {
					return err
				}
				continue
			}

			plaintext, err := security.DecryptAESGCM(enrollments[i].EncryptedData, oldAESKey)
			if err != nil {
				return fmt.Errorf("biometric.RotateKey decrypt enrollment %s: %w", enrollments[i].ID, err)
			}
			reencrypted, err := security.EncryptAESGCM(plaintext, newAESKey)
			if err != nil {
				return fmt.Errorf("biometric.RotateKey re-encrypt enrollment %s: %w", enrollments[i].ID, err)
			}

			enrollments[i].EncryptedData = reencrypted
			enrollments[i].EncryptionKeyID = newKey.ID
			enrollments[i].UpdatedAt = time.Now().UTC()
			if err := s.biometricRepo.Update(txCtx, &enrollments[i]); err != nil {
				return err
			}
		}

		if oldKey != nil && err != domain.ErrNotFound {
			now := time.Now().UTC()
			oldKey.Status = domain.EncryptionKeyStatusRotated
			oldKey.RotatedAt = &now
			if err := s.encKeyRepo.Update(txCtx, oldKey); err != nil {
				return err
			}
		}

		return s.auditSvc.Log(txCtx, "encryption_key.rotated", "encryption_key", newKey.ID, performedBy, map[string]interface{}{
			"re_encrypted_enrollments": len(enrollments),
		})
	})
	if err != nil {
		return nil, err
	}

	return newKey, nil
}

// GetActiveKey returns the currently active biometric encryption key.
func (s *BiometricServiceImpl) GetActiveKey(ctx context.Context) (*domain.EncryptionKey, error) {
	return s.encKeyRepo.GetActive(ctx, "biometric")
}

// ListKeys returns all encryption keys for the biometric purpose.
func (s *BiometricServiceImpl) ListKeys(ctx context.Context) ([]domain.EncryptionKey, error) {
	return s.encKeyRepo.List(ctx, "biometric")
}

func (s *BiometricServiceImpl) ensureActiveKey(ctx context.Context) (*domain.EncryptionKey, error) {
	activeKey, err := s.encKeyRepo.GetActive(ctx, "biometric")
	if errors.Is(err, domain.ErrNotFound) || (err == nil && len(activeKey.WrappedDEK) == 0) {
		// No active key, or the active key pre-dates envelope encryption (legacy row).
		// Rotate to create a proper key with a wrapped DEK.
		if _, rotateErr := s.RotateKey(ctx, domain.SystemActorID); rotateErr != nil {
			return nil, rotateErr
		}
		return s.encKeyRepo.GetActive(ctx, "biometric")
	}
	if err != nil {
		return nil, err
	}
	return activeKey, nil
}

// Compile-time interface assertion.
var _ BiometricService = (*BiometricServiceImpl)(nil)
