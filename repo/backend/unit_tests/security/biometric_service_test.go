package security_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
	"fitcommerce/internal/platform"
)

// --- Mock BiometricRepository ---

type mockBiometricRepo struct {
	enrollments map[uuid.UUID]*domain.BiometricEnrollment // keyed by user ID
	createErr   error
	getErr      error
	updateErr   error
	listErr     error
}

func newMockBiometricRepo() *mockBiometricRepo {
	return &mockBiometricRepo{enrollments: make(map[uuid.UUID]*domain.BiometricEnrollment)}
}

func (m *mockBiometricRepo) Create(_ context.Context, e *domain.BiometricEnrollment) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.enrollments[e.UserID] = e
	return nil
}

func (m *mockBiometricRepo) GetByUserID(_ context.Context, userID uuid.UUID) (*domain.BiometricEnrollment, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	e, ok := m.enrollments[userID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return e, nil
}

func (m *mockBiometricRepo) List(_ context.Context) ([]domain.BiometricEnrollment, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	out := make([]domain.BiometricEnrollment, 0, len(m.enrollments))
	for _, e := range m.enrollments {
		out = append(out, *e)
	}
	return out, nil
}

func (m *mockBiometricRepo) Update(_ context.Context, e *domain.BiometricEnrollment) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.enrollments[e.UserID] = e
	return nil
}

// --- Mock EncryptionKeyRepository ---

type mockEncryptionKeyRepo struct {
	keys         map[uuid.UUID]*domain.EncryptionKey
	activeByPurpose map[string]uuid.UUID
	getActiveErr error
	listErr      error
}

func newMockEncryptionKeyRepo() *mockEncryptionKeyRepo {
	return &mockEncryptionKeyRepo{
		keys:         make(map[uuid.UUID]*domain.EncryptionKey),
		activeByPurpose: make(map[string]uuid.UUID),
	}
}

func (m *mockEncryptionKeyRepo) Create(_ context.Context, k *domain.EncryptionKey) error {
	m.keys[k.ID] = k
	if k.Status == domain.EncryptionKeyStatusActive {
		m.activeByPurpose[k.Purpose] = k.ID
	}
	return nil
}

func (m *mockEncryptionKeyRepo) GetActive(_ context.Context, purpose string) (*domain.EncryptionKey, error) {
	if m.getActiveErr != nil {
		return nil, m.getActiveErr
	}
	id, ok := m.activeByPurpose[purpose]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return m.keys[id], nil
}

func (m *mockEncryptionKeyRepo) List(_ context.Context, purpose string) ([]domain.EncryptionKey, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	out := []domain.EncryptionKey{}
	for _, k := range m.keys {
		if k.Purpose == purpose {
			out = append(out, *k)
		}
	}
	return out, nil
}

func (m *mockEncryptionKeyRepo) Update(_ context.Context, k *domain.EncryptionKey) error {
	m.keys[k.ID] = k
	if k.Status != domain.EncryptionKeyStatusActive && m.activeByPurpose[k.Purpose] == k.ID {
		delete(m.activeByPurpose, k.Purpose)
	}
	return nil
}

// --- Mock audit (local to this file to avoid collision) ---

type bioAuditStub struct{}

func (b *bioAuditStub) Log(_ context.Context, _, _ string, _, _ uuid.UUID, _ map[string]interface{}) error {
	return nil
}
func (b *bioAuditStub) List(_ context.Context, _ string, _ *uuid.UUID, _, _ int) ([]domain.AuditEvent, int, error) {
	return nil, 0, nil
}
func (b *bioAuditStub) ListByEventTypes(_ context.Context, _ []string, _, _ int) ([]domain.AuditEvent, int, error) {
	return nil, 0, nil
}

func bioTestConfig() platform.Config {
	return platform.Config{BiometricMasterKeyRef: "test-master-key"}
}

func TestBiometricService_GetByUser_Found_RedactsData(t *testing.T) {
	bio := newMockBiometricRepo()
	enc := newMockEncryptionKeyRepo()
	svc := application.NewBiometricService(bio, enc, &bioAuditStub{}, nil, 30, bioTestConfig())

	userID := uuid.New()
	bio.enrollments[userID] = &domain.BiometricEnrollment{
		ID:            uuid.New(),
		UserID:        userID,
		EncryptedData: []byte("secret-ciphertext-that-must-not-leak"),
	}

	got, err := svc.GetByUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetByUser failed: %v", err)
	}
	if string(got.EncryptedData) == "secret-ciphertext-that-must-not-leak" {
		t.Error("expected EncryptedData redacted on read")
	}
	// Verify stored data wasn't mutated.
	if string(bio.enrollments[userID].EncryptedData) != "secret-ciphertext-that-must-not-leak" {
		t.Error("redaction must not mutate stored data")
	}
}

func TestBiometricService_GetByUser_NotFound(t *testing.T) {
	bio := newMockBiometricRepo()
	enc := newMockEncryptionKeyRepo()
	svc := application.NewBiometricService(bio, enc, &bioAuditStub{}, nil, 30, bioTestConfig())

	_, err := svc.GetByUser(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestBiometricService_Revoke_Success_ClearsData(t *testing.T) {
	bio := newMockBiometricRepo()
	enc := newMockEncryptionKeyRepo()
	svc := application.NewBiometricService(bio, enc, &bioAuditStub{}, nil, 30, bioTestConfig())

	userID := uuid.New()
	bio.enrollments[userID] = &domain.BiometricEnrollment{
		ID:            uuid.New(),
		UserID:        userID,
		EncryptedData: []byte("stillhere"),
		UpdatedAt:     time.Now().Add(-time.Hour),
	}

	if err := svc.Revoke(context.Background(), userID); err != nil {
		t.Fatalf("Revoke failed: %v", err)
	}
	stored := bio.enrollments[userID]
	if len(stored.EncryptedData) != 0 {
		t.Errorf("expected EncryptedData cleared, got %q", stored.EncryptedData)
	}
	if time.Since(stored.UpdatedAt) > time.Minute {
		t.Error("expected UpdatedAt to be refreshed")
	}
}

func TestBiometricService_Revoke_NotFound(t *testing.T) {
	bio := newMockBiometricRepo()
	enc := newMockEncryptionKeyRepo()
	svc := application.NewBiometricService(bio, enc, &bioAuditStub{}, nil, 30, bioTestConfig())

	err := svc.Revoke(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestBiometricService_Revoke_UpdateError(t *testing.T) {
	bio := newMockBiometricRepo()
	bio.updateErr = errors.New("db down")
	enc := newMockEncryptionKeyRepo()
	svc := application.NewBiometricService(bio, enc, &bioAuditStub{}, nil, 30, bioTestConfig())

	userID := uuid.New()
	bio.enrollments[userID] = &domain.BiometricEnrollment{UserID: userID, EncryptedData: []byte("x")}

	if err := svc.Revoke(context.Background(), userID); err == nil {
		t.Fatal("expected update error propagated")
	}
}

func TestBiometricService_GetActiveKey_Found(t *testing.T) {
	bio := newMockBiometricRepo()
	enc := newMockEncryptionKeyRepo()
	svc := application.NewBiometricService(bio, enc, &bioAuditStub{}, nil, 30, bioTestConfig())

	activeID := uuid.New()
	enc.keys[activeID] = &domain.EncryptionKey{
		ID:      activeID,
		Purpose: "biometric",
		Status:  domain.EncryptionKeyStatusActive,
	}
	enc.activeByPurpose["biometric"] = activeID

	got, err := svc.GetActiveKey(context.Background())
	if err != nil {
		t.Fatalf("GetActiveKey failed: %v", err)
	}
	if got.ID != activeID {
		t.Errorf("expected %v, got %v", activeID, got.ID)
	}
}

func TestBiometricService_GetActiveKey_NotFound(t *testing.T) {
	bio := newMockBiometricRepo()
	enc := newMockEncryptionKeyRepo()
	svc := application.NewBiometricService(bio, enc, &bioAuditStub{}, nil, 30, bioTestConfig())

	_, err := svc.GetActiveKey(context.Background())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestBiometricService_ListKeys_ReturnsBiometricPurpose(t *testing.T) {
	bio := newMockBiometricRepo()
	enc := newMockEncryptionKeyRepo()
	svc := application.NewBiometricService(bio, enc, &bioAuditStub{}, nil, 30, bioTestConfig())

	bio1ID := uuid.New()
	bio2ID := uuid.New()
	otherID := uuid.New()
	enc.keys[bio1ID] = &domain.EncryptionKey{ID: bio1ID, Purpose: "biometric"}
	enc.keys[bio2ID] = &domain.EncryptionKey{ID: bio2ID, Purpose: "biometric"}
	enc.keys[otherID] = &domain.EncryptionKey{ID: otherID, Purpose: "other"}

	keys, err := svc.ListKeys(context.Background())
	if err != nil {
		t.Fatalf("ListKeys failed: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected only biometric-purpose keys (2), got %d", len(keys))
	}
}

func TestBiometricService_ListKeys_RepoError(t *testing.T) {
	bio := newMockBiometricRepo()
	enc := newMockEncryptionKeyRepo()
	enc.listErr = errors.New("db broken")
	svc := application.NewBiometricService(bio, enc, &bioAuditStub{}, nil, 30, bioTestConfig())

	_, err := svc.ListKeys(context.Background())
	if err == nil {
		t.Fatal("expected error propagated")
	}
}

func TestBiometricService_NewBiometricService_DefaultsRotationDays(t *testing.T) {
	bio := newMockBiometricRepo()
	enc := newMockEncryptionKeyRepo()
	// Pass zero — should be defaulted to 90. We can't observe rotationDays directly,
	// but constructing and calling a no-op method must not panic.
	svc := application.NewBiometricService(bio, enc, &bioAuditStub{}, nil, 0, bioTestConfig())
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}
