package domain

import (
	"time"

	"github.com/google/uuid"
)

// BiometricEnrollment represents a user's biometric enrollment data,
// stored in encrypted form.
type BiometricEnrollment struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	EncryptedData   []byte
	EncryptionKeyID uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// EncryptionKey represents a key used for encrypting biometric data.
type EncryptionKey struct {
	ID           uuid.UUID
	KeyReference string
	Purpose      string // "biometric"
	Status       EncryptionKeyStatus
	ActivatedAt  time.Time
	RotatedAt    *time.Time
	ExpiresAt    time.Time
}

// NeedsRotation returns true if the number of days since the key was activated
// meets or exceeds the specified rotation interval.
func (k *EncryptionKey) NeedsRotation(rotationDays int) bool {
	daysSinceActivation := time.Since(k.ActivatedAt).Hours() / 24
	return daysSinceActivation >= float64(rotationDays)
}
