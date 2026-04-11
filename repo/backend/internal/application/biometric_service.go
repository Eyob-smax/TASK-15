package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
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
}

// NewBiometricService creates a BiometricServiceImpl backed by the given repositories.
func NewBiometricService(
	biometricRepo store.BiometricRepository,
	encKeyRepo store.EncryptionKeyRepository,
	auditSvc AuditService,
	txPool *pgxpool.Pool,
	rotationDays int,
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
	}
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

	aesKey, err := security.DeriveKeyFromRef(activeKey.KeyReference)
	if err != nil {
		return nil, fmt.Errorf("biometric.Register: key derivation failed: %w", err)
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

		newKey = &domain.EncryptionKey{
			ID:           uuid.New(),
			KeyReference: uuid.New().String(),
			Purpose:      "biometric",
			Status:       domain.EncryptionKeyStatusActive,
			ActivatedAt:  time.Now().UTC(),
			ExpiresAt:    time.Now().UTC().AddDate(0, 0, s.rotationDays),
		}

		var oldAESKey []byte
		if oldKey != nil && err != domain.ErrNotFound {
			oldAESKey, err = security.DeriveKeyFromRef(oldKey.KeyReference)
			if err != nil {
				return fmt.Errorf("biometric.RotateKey old key derivation failed: %w", err)
			}
		}

		newAESKey, err := security.DeriveKeyFromRef(newKey.KeyReference)
		if err != nil {
			return fmt.Errorf("biometric.RotateKey new key derivation failed: %w", err)
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
	if errors.Is(err, domain.ErrNotFound) {
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
