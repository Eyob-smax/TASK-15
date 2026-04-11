package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
)

// ---------------------------------------------------------------------------
// BiometricStore
// ---------------------------------------------------------------------------

// BiometricStore implements store.BiometricRepository using a pgx connection pool.
type BiometricStore struct {
	pool *pgxpool.Pool
}

// NewBiometricStore creates a new BiometricStore backed by the given connection pool.
func NewBiometricStore(pool *pgxpool.Pool) *BiometricStore {
	return &BiometricStore{pool: pool}
}

// Create inserts a new biometric enrollment record.
func (s *BiometricStore) Create(ctx context.Context, enrollment *domain.BiometricEnrollment) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		INSERT INTO biometric_enrollments (id, user_id, encrypted_data, encryption_key_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := db.Exec(ctx, q,
		enrollment.ID,
		enrollment.UserID,
		enrollment.EncryptedData,
		enrollment.EncryptionKeyID,
		enrollment.CreatedAt,
		enrollment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("biometric_store.Create: %w", err)
	}
	return nil
}

// GetByUserID returns the biometric enrollment for the given user.
// Returns domain.ErrNotFound if no matching record exists.
func (s *BiometricStore) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.BiometricEnrollment, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `
		SELECT id, user_id, encrypted_data, encryption_key_id, created_at, updated_at
		FROM biometric_enrollments
		WHERE user_id = $1
		LIMIT 1`

	row := db.QueryRow(ctx, q, userID)
	enrollment, err := scanBiometricEnrollment(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("biometric_store.GetByUserID: %w", err)
	}
	return enrollment, nil
}

// List returns all biometric enrollments, used during key rotation to
// re-encrypt active templates under the new envelope key.
func (s *BiometricStore) List(ctx context.Context) ([]domain.BiometricEnrollment, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `
		SELECT id, user_id, encrypted_data, encryption_key_id, created_at, updated_at
		FROM biometric_enrollments
		ORDER BY created_at ASC`

	rows, err := db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("biometric_store.List query: %w", err)
	}
	defer rows.Close()

	var enrollments []domain.BiometricEnrollment
	for rows.Next() {
		enrollment, err := scanBiometricEnrollment(rows)
		if err != nil {
			return nil, fmt.Errorf("biometric_store.List scan: %w", err)
		}
		enrollments = append(enrollments, *enrollment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("biometric_store.List rows: %w", err)
	}
	return enrollments, nil
}

// Update modifies the encrypted_data, encryption_key_id, and updated_at of an
// existing biometric enrollment. updated_at is always set to the current UTC time.
func (s *BiometricStore) Update(ctx context.Context, enrollment *domain.BiometricEnrollment) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		UPDATE biometric_enrollments
		SET encrypted_data = $1, encryption_key_id = $2, updated_at = $3
		WHERE id = $4`

	_, err := db.Exec(ctx, q,
		enrollment.EncryptedData,
		enrollment.EncryptionKeyID,
		time.Now().UTC(),
		enrollment.ID,
	)
	if err != nil {
		return fmt.Errorf("biometric_store.Update: %w", err)
	}
	return nil
}

// scanBiometricEnrollment reads a single biometric_enrollments row into a domain struct.
func scanBiometricEnrollment(row pgx.Row) (*domain.BiometricEnrollment, error) {
	var e domain.BiometricEnrollment
	err := row.Scan(
		&e.ID,
		&e.UserID,
		&e.EncryptedData,
		&e.EncryptionKeyID,
		&e.CreatedAt,
		&e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// Compile-time interface assertion.
var _ store.BiometricRepository = (*BiometricStore)(nil)

// ---------------------------------------------------------------------------
// EncryptionKeyStore
// ---------------------------------------------------------------------------

// EncryptionKeyStore implements store.EncryptionKeyRepository using a pgx connection pool.
type EncryptionKeyStore struct {
	pool *pgxpool.Pool
}

// NewEncryptionKeyStore creates a new EncryptionKeyStore backed by the given connection pool.
func NewEncryptionKeyStore(pool *pgxpool.Pool) *EncryptionKeyStore {
	return &EncryptionKeyStore{pool: pool}
}

// Create inserts a new encryption key record. created_at is set to the current
// UTC time in Go because the domain struct does not carry that field.
func (s *EncryptionKeyStore) Create(ctx context.Context, key *domain.EncryptionKey) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		INSERT INTO encryption_keys (id, key_reference, purpose, status, activated_at, rotated_at, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := db.Exec(ctx, q,
		key.ID,
		key.KeyReference,
		key.Purpose,
		key.Status,
		key.ActivatedAt,
		key.RotatedAt,
		key.ExpiresAt,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("encryption_key_store.Create: %w", err)
	}
	return nil
}

// GetActive returns the most recently activated key for the given purpose whose
// status is 'active'. Returns domain.ErrNotFound if no such key exists.
func (s *EncryptionKeyStore) GetActive(ctx context.Context, purpose string) (*domain.EncryptionKey, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `
		SELECT id, key_reference, purpose, status, activated_at, rotated_at, expires_at
		FROM encryption_keys
		WHERE purpose = $1 AND status = 'active'
		ORDER BY activated_at DESC
		LIMIT 1`

	row := db.QueryRow(ctx, q, purpose)
	key, err := scanEncryptionKey(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("encryption_key_store.GetActive: %w", err)
	}
	return key, nil
}

// List returns all encryption keys for the given purpose ordered by activation
// time descending.
func (s *EncryptionKeyStore) List(ctx context.Context, purpose string) ([]domain.EncryptionKey, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `
		SELECT id, key_reference, purpose, status, activated_at, rotated_at, expires_at
		FROM encryption_keys
		WHERE purpose = $1
		ORDER BY activated_at DESC`

	rows, err := db.Query(ctx, q, purpose)
	if err != nil {
		return nil, fmt.Errorf("encryption_key_store.List query: %w", err)
	}
	defer rows.Close()

	var keys []domain.EncryptionKey
	for rows.Next() {
		key, err := scanEncryptionKey(rows)
		if err != nil {
			return nil, fmt.Errorf("encryption_key_store.List scan: %w", err)
		}
		keys = append(keys, *key)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("encryption_key_store.List rows: %w", err)
	}
	return keys, nil
}

// Update modifies the status and rotated_at of an existing encryption key.
func (s *EncryptionKeyStore) Update(ctx context.Context, key *domain.EncryptionKey) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		UPDATE encryption_keys
		SET status = $1, rotated_at = $2
		WHERE id = $3`

	_, err := db.Exec(ctx, q,
		key.Status,
		key.RotatedAt,
		key.ID,
	)
	if err != nil {
		return fmt.Errorf("encryption_key_store.Update: %w", err)
	}
	return nil
}

// scanEncryptionKey reads a single encryption_keys row into a domain struct.
// The status column is scanned as a plain string and cast to domain.EncryptionKeyStatus.
// created_at is present in the DB but absent from the domain struct, so it is not scanned.
func scanEncryptionKey(row pgx.Row) (*domain.EncryptionKey, error) {
	var k domain.EncryptionKey
	var statusStr string
	err := row.Scan(
		&k.ID,
		&k.KeyReference,
		&k.Purpose,
		&statusStr,
		&k.ActivatedAt,
		&k.RotatedAt,
		&k.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}
	k.Status = domain.EncryptionKeyStatus(statusStr)
	return &k, nil
}

// Compile-time interface assertion.
var _ store.EncryptionKeyRepository = (*EncryptionKeyStore)(nil)
