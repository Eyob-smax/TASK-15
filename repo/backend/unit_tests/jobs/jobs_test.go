package jobs_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
	"fitcommerce/internal/jobs"
)

// --- Mock services ---

type mockOrderService struct {
	autoCloseCalled int
	autoCloseCount  int
	autoCloseErr    error
}

func (m *mockOrderService) AutoCloseExpired(_ context.Context, _ time.Time) (int, error) {
	m.autoCloseCalled++
	return m.autoCloseCount, m.autoCloseErr
}

// Stub all other OrderService methods.
func (m *mockOrderService) Create(_ context.Context, _ *domain.Order) (*domain.Order, error) {
	return nil, nil
}
func (m *mockOrderService) Get(_ context.Context, _ uuid.UUID) (*domain.Order, error) {
	return nil, nil
}
func (m *mockOrderService) List(_ context.Context, _ *uuid.UUID, _, _ int) ([]domain.Order, int, error) {
	return nil, 0, nil
}
func (m *mockOrderService) Pay(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID) error { return nil }
func (m *mockOrderService) Cancel(_ context.Context, _ uuid.UUID, _ uuid.UUID) error { return nil }
func (m *mockOrderService) Refund(_ context.Context, _ uuid.UUID, _ uuid.UUID) error { return nil }
func (m *mockOrderService) AddNote(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID) error {
	return nil
}
func (m *mockOrderService) GetForActor(_ context.Context, _ *domain.User, _ uuid.UUID) (*domain.Order, error) {
	return nil, nil
}
func (m *mockOrderService) ListForActor(_ context.Context, _ *domain.User, _, _ int) ([]domain.Order, int, error) {
	return nil, 0, nil
}
func (m *mockOrderService) CancelForActor(_ context.Context, _ *domain.User, _ uuid.UUID) error {
	return nil
}
func (m *mockOrderService) Split(_ context.Context, _ uuid.UUID, _ []int, _ *application.FulfillmentInput) ([]domain.Order, error) {
	return nil, nil
}
func (m *mockOrderService) SplitForActor(_ context.Context, _ *domain.User, _ uuid.UUID, _ []int, _ *application.FulfillmentInput) ([]domain.Order, error) {
	return nil, nil
}
func (m *mockOrderService) Merge(_ context.Context, _ []uuid.UUID, _ *application.FulfillmentInput) (*domain.Order, error) {
	return nil, nil
}
func (m *mockOrderService) MergeForActor(_ context.Context, _ *domain.User, _ []uuid.UUID, _ *application.FulfillmentInput) (*domain.Order, error) {
	return nil, nil
}
func (m *mockOrderService) GetTimeline(_ context.Context, _ uuid.UUID) ([]domain.OrderTimelineEntry, error) {
	return nil, nil
}
func (m *mockOrderService) GetTimelineForActor(_ context.Context, _ *domain.User, _ uuid.UUID) ([]domain.OrderTimelineEntry, error) {
	return nil, nil
}

type mockCampaignService struct {
	listPastCutoffCalls  int
	evaluateCalls        int
	campaignsToReturn    []domain.GroupBuyCampaign
}

func (m *mockCampaignService) ListPastCutoff(_ context.Context, _ time.Time) ([]domain.GroupBuyCampaign, error) {
	m.listPastCutoffCalls++
	return m.campaignsToReturn, nil
}
func (m *mockCampaignService) EvaluateAtCutoff(_ context.Context, _ uuid.UUID, _ time.Time) error {
	m.evaluateCalls++
	return nil
}

// Stub all other CampaignService methods.
func (m *mockCampaignService) Create(_ context.Context, _ *domain.GroupBuyCampaign) (*domain.GroupBuyCampaign, error) {
	return nil, nil
}
func (m *mockCampaignService) Get(_ context.Context, _ uuid.UUID) (*domain.GroupBuyCampaign, error) {
	return nil, nil
}
func (m *mockCampaignService) List(_ context.Context, _, _ int) ([]domain.GroupBuyCampaign, int, error) {
	return nil, 0, nil
}
func (m *mockCampaignService) Join(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ int) (*domain.GroupBuyParticipant, error) {
	return nil, nil
}
func (m *mockCampaignService) Cancel(_ context.Context, _ uuid.UUID) error { return nil }

// --- Tests ---

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestAutoCloseJob_TickCallsAutoCloseExpired(t *testing.T) {
	mockSvc := &mockOrderService{autoCloseCount: 3}
	job := jobs.NewAutoCloseJobWithInterval(mockSvc, testLogger(), 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	job.Run(ctx)

	if mockSvc.autoCloseCalled == 0 {
		t.Error("expected AutoCloseExpired to be called at least once")
	}
}

func TestAutoCloseJob_ContextCancel_Exits(t *testing.T) {
	mockSvc := &mockOrderService{}
	job := jobs.NewAutoCloseJobWithInterval(mockSvc, testLogger(), 1*time.Second) // long interval

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		job.Run(ctx)
		close(done)
	}()

	cancel()
	select {
	case <-done:
		// good — job exited
	case <-time.After(200 * time.Millisecond):
		t.Error("expected job to exit after context cancel")
	}
}

func TestCutoffEvalJob_TickCallsEvaluate(t *testing.T) {
	mockSvc := &mockCampaignService{
		campaignsToReturn: []domain.GroupBuyCampaign{
			{ID: uuid.New()},
			{ID: uuid.New()},
		},
	}
	job := jobs.NewCutoffEvalJobWithInterval(mockSvc, testLogger(), 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	job.Run(ctx)

	if mockSvc.listPastCutoffCalls == 0 {
		t.Error("expected ListPastCutoff to be called")
	}
	if mockSvc.evaluateCalls < 2 {
		t.Errorf("expected at least 2 EvaluateAtCutoff calls (one per campaign), got %d", mockSvc.evaluateCalls)
	}
}

func TestCutoffEvalJob_ContextCancel_Exits(t *testing.T) {
	mockSvc := &mockCampaignService{}
	job := jobs.NewCutoffEvalJobWithInterval(mockSvc, testLogger(), 1*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		job.Run(ctx)
		close(done)
	}()

	cancel()
	select {
	case <-done:
		// good
	case <-time.After(200 * time.Millisecond):
		t.Error("expected job to exit after context cancel")
	}
}

// --- Mock services for procurement/backup/retention jobs ---

type mockBackupService struct {
	triggerCalled int
	triggerErr    error
}

func (m *mockBackupService) Trigger(_ context.Context, _ *uuid.UUID) (*domain.BackupRun, error) {
	m.triggerCalled++
	if m.triggerErr != nil {
		return nil, m.triggerErr
	}
	return &domain.BackupRun{ID: uuid.New(), Status: domain.BackupStatusCompleted}, nil
}

func (m *mockBackupService) GetByID(_ context.Context, _ uuid.UUID) (*domain.BackupRun, error) {
	return nil, nil
}

func (m *mockBackupService) List(_ context.Context, _, _ int) ([]domain.BackupRun, int, error) {
	return nil, 0, nil
}

type mockVarianceService struct {
	listCalled     int
	escalateCalled int
	escalateCount  int
	escalateErr    error
	records        []domain.VarianceRecord
}

func (m *mockVarianceService) Get(_ context.Context, _ uuid.UUID) (*domain.VarianceRecord, error) {
	return nil, nil
}

func (m *mockVarianceService) List(_ context.Context, _ *domain.VarianceStatus, _, _ int) ([]domain.VarianceRecord, int, error) {
	m.listCalled++
	return m.records, len(m.records), nil
}

func (m *mockVarianceService) Resolve(_ context.Context, _ uuid.UUID, _ string, _ string, _ *int, _ uuid.UUID) error {
	return nil
}

func (m *mockVarianceService) EscalateOverdue(_ context.Context) (int, error) {
	m.escalateCalled++
	return m.escalateCount, m.escalateErr
}

type mockRetentionService struct {
	cleanupCalled int
	cleanupErr    error
}

func (m *mockRetentionService) GetByEntityType(_ context.Context, _ string) (*domain.RetentionPolicy, error) {
	return nil, nil
}

func (m *mockRetentionService) List(_ context.Context) ([]domain.RetentionPolicy, error) {
	return nil, nil
}

func (m *mockRetentionService) Update(_ context.Context, _ *domain.RetentionPolicy) error {
	return nil
}

func (m *mockRetentionService) RunCleanup(_ context.Context) error {
	m.cleanupCalled++
	return m.cleanupErr
}

type mockBiometricService struct {
	getActiveCalled int
	rotateCalled    int
	activeKey       *domain.EncryptionKey
	getActiveErr    error
	rotateErr       error
}

func (m *mockBiometricService) Register(_ context.Context, _ uuid.UUID, _ string) (*domain.BiometricEnrollment, error) {
	return nil, nil
}

func (m *mockBiometricService) GetByUser(_ context.Context, _ uuid.UUID) (*domain.BiometricEnrollment, error) {
	return nil, nil
}

func (m *mockBiometricService) Revoke(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockBiometricService) RotateKey(_ context.Context, _ uuid.UUID) (*domain.EncryptionKey, error) {
	m.rotateCalled++
	if m.rotateErr != nil {
		return nil, m.rotateErr
	}
	return &domain.EncryptionKey{ID: uuid.New(), Purpose: "biometric", Status: domain.EncryptionKeyStatusActive}, nil
}

func (m *mockBiometricService) GetActiveKey(_ context.Context) (*domain.EncryptionKey, error) {
	m.getActiveCalled++
	if m.getActiveErr != nil {
		return nil, m.getActiveErr
	}
	return m.activeKey, nil
}

func (m *mockBiometricService) ListKeys(_ context.Context) ([]domain.EncryptionKey, error) {
	return nil, nil
}

// --- Tests for procurement/backup/retention jobs ---

func TestBackupJob_TickTrigger(t *testing.T) {
	mockSvc := &mockBackupService{}
	job := jobs.NewBackupJobWithInterval(mockSvc, testLogger(), 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	job.Run(ctx)

	if mockSvc.triggerCalled == 0 {
		t.Error("expected BackupService.Trigger to be called at least once")
	}
}

func TestVarianceDeadlineJob_OverdueLogsWarn(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-25 * time.Hour)
	overdueRecord := domain.VarianceRecord{
		ID:                uuid.New(),
		POLineID:          uuid.New(),
		Type:              domain.VarianceTypeShortage,
		Status:            domain.VarianceStatusOpen,
		ResolutionDueDate: yesterday,
		CreatedAt:         now.Add(-48 * time.Hour),
	}
	mockSvc := &mockVarianceService{records: []domain.VarianceRecord{overdueRecord}}
	job := jobs.NewVarianceDeadlineJobWithInterval(mockSvc, testLogger(), 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	job.Run(ctx)

	if mockSvc.escalateCalled == 0 {
		t.Error("expected VarianceService.EscalateOverdue to be called at least once")
	}
}

func TestRetentionCleanupJob_TickRunsCleanup(t *testing.T) {
	mockSvc := &mockRetentionService{}
	job := jobs.NewRetentionCleanupJobWithInterval(mockSvc, testLogger(), 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	job.Run(ctx)

	if mockSvc.cleanupCalled == 0 {
		t.Error("expected RetentionService.RunCleanup to be called at least once")
	}
}

func TestBiometricKeyRotationJob_BootstrapsMissingKey(t *testing.T) {
	mockSvc := &mockBiometricService{getActiveErr: domain.ErrNotFound}
	job := jobs.NewBiometricKeyRotationJobWithInterval(mockSvc, testLogger(), 90, 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	job.Run(ctx)

	if mockSvc.getActiveCalled == 0 {
		t.Error("expected BiometricService.GetActiveKey to be called at least once")
	}
	if mockSvc.rotateCalled == 0 {
		t.Error("expected missing active key to trigger rotation/bootstrap")
	}
}

func TestBiometricKeyRotationJob_RotatesDueKey(t *testing.T) {
	now := time.Now().UTC()
	mockSvc := &mockBiometricService{
		activeKey: &domain.EncryptionKey{
			ID:          uuid.New(),
			Purpose:     "biometric",
			Status:      domain.EncryptionKeyStatusActive,
			ActivatedAt: now.AddDate(0, 0, -91),
			ExpiresAt:   now.Add(-1 * time.Hour),
		},
	}
	job := jobs.NewBiometricKeyRotationJobWithInterval(mockSvc, testLogger(), 90, 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	job.Run(ctx)

	if mockSvc.rotateCalled == 0 {
		t.Error("expected due biometric key to rotate")
	}
}

func TestBiometricKeyRotationJob_SkipsFreshKey(t *testing.T) {
	now := time.Now().UTC()
	mockSvc := &mockBiometricService{
		activeKey: &domain.EncryptionKey{
			ID:          uuid.New(),
			Purpose:     "biometric",
			Status:      domain.EncryptionKeyStatusActive,
			ActivatedAt: now.AddDate(0, 0, -7),
			ExpiresAt:   now.AddDate(0, 0, 30),
		},
	}
	job := jobs.NewBiometricKeyRotationJobWithInterval(mockSvc, testLogger(), 90, time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		job.Run(ctx)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()
	<-done

	if mockSvc.rotateCalled != 0 {
		t.Error("expected fresh biometric key to skip rotation")
	}
}
