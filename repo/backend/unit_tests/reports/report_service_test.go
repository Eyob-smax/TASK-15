package reports_test

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

// --- Mock repositories ---

type mockReportRepo struct {
	reports map[uuid.UUID]*domain.ReportDefinition
}

func newMockReportRepo() *mockReportRepo {
	report := &domain.ReportDefinition{
		ID:           uuid.New(),
		Name:         "Test Report",
		ReportType:   "items_summary",
		AllowedRoles: []domain.UserRole{domain.UserRoleAdministrator},
		CreatedAt:    time.Now(),
	}
	return &mockReportRepo{reports: map[uuid.UUID]*domain.ReportDefinition{report.ID: report}}
}

func (m *mockReportRepo) Create(_ context.Context, report *domain.ReportDefinition) error {
	m.reports[report.ID] = report
	return nil
}

func (m *mockReportRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.ReportDefinition, error) {
	report, ok := m.reports[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return report, nil
}

func (m *mockReportRepo) List(_ context.Context) ([]domain.ReportDefinition, error) {
	list := make([]domain.ReportDefinition, 0, len(m.reports))
	for _, v := range m.reports {
		list = append(list, *v)
	}
	return list, nil
}

type mockExportRepo struct {
	jobs map[uuid.UUID]*domain.ExportJob
}

func newMockExportRepo() *mockExportRepo {
	return &mockExportRepo{jobs: make(map[uuid.UUID]*domain.ExportJob)}
}

func (m *mockExportRepo) Create(_ context.Context, job *domain.ExportJob) error {
	m.jobs[job.ID] = job
	return nil
}

func (m *mockExportRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.ExportJob, error) {
	job, ok := m.jobs[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return job, nil
}

func (m *mockExportRepo) Update(_ context.Context, job *domain.ExportJob) error {
	if _, ok := m.jobs[job.ID]; !ok {
		return domain.ErrNotFound
	}
	m.jobs[job.ID] = job
	return nil
}

// --- Mock audit service ---

type mockAuditService struct{}

func (m *mockAuditService) Log(_ context.Context, _, _ string, _, _ uuid.UUID, _ map[string]interface{}) error {
	return nil
}

func (m *mockAuditService) List(_ context.Context, _ string, _ *uuid.UUID, _, _ int) ([]domain.AuditEvent, int, error) {
	return nil, 0, nil
}

func (m *mockAuditService) ListByEventTypes(_ context.Context, _ []string, _, _ int) ([]domain.AuditEvent, int, error) {
	return nil, 0, nil
}

// --- Helper ---

func newReportService(reportRepo *mockReportRepo, exportRepo *mockExportRepo, exportPath string) application.ReportService {
	cfg := platform.Config{ExportPath: exportPath}
	return application.NewReportService(reportRepo, exportRepo, cfg, &mockAuditService{}, nil)
}

// --- Tests ---

func TestReportList_RoleFilter(t *testing.T) {
	reportRepo := newMockReportRepo()
	exportRepo := newMockExportRepo()
	svc := newReportService(reportRepo, exportRepo, t.TempDir())

	t.Run("admin gets admin reports", func(t *testing.T) {
		reports, err := svc.List(context.Background(), domain.UserRoleAdministrator)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(reports) == 0 {
			t.Error("expected administrator to see at least 1 report")
		}
	})

	t.Run("coach does not get admin-only reports", func(t *testing.T) {
		reports, err := svc.List(context.Background(), domain.UserRoleCoach)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		// The seeded report only allows administrator, so coach should see 0.
		if len(reports) != 0 {
			t.Errorf("expected coach to see 0 admin-only reports, got %d", len(reports))
		}
	})
}

func TestGetData_ReturnsStructuredData(t *testing.T) {
	reportRepo := newMockReportRepo()
	exportRepo := newMockExportRepo()
	svc := newReportService(reportRepo, exportRepo, t.TempDir())

	// Pick the one seeded report.
	var reportID uuid.UUID
	for id := range reportRepo.reports {
		reportID = id
	}

	data, err := svc.GetData(context.Background(), reportID, nil, domain.UserRoleAdministrator)
	if err != nil {
		t.Fatalf("GetData failed: %v", err)
	}
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", data)
	}
	if dataMap["report_type"] == nil {
		t.Error("expected report_type field in GetData result")
	}
}

func TestGenerateExport_CSV_CreatesFile(t *testing.T) {
	reportRepo := newMockReportRepo()
	exportRepo := newMockExportRepo()
	exportPath := t.TempDir()
	svc := newReportService(reportRepo, exportRepo, exportPath)

	var reportID uuid.UUID
	for id := range reportRepo.reports {
		reportID = id
	}

	job, err := svc.GenerateExport(context.Background(), reportID, domain.ExportFormatCSV, map[string]string{"location_id": "loc-1"}, uuid.New(), domain.UserRoleAdministrator)
	if err != nil {
		t.Fatalf("GenerateExport CSV failed: %v", err)
	}
	if job.Status != domain.ExportStatusCompleted {
		t.Errorf("expected status=completed, got %v", job.Status)
	}
	if job.FilePath == "" {
		t.Error("expected non-empty FilePath for completed CSV export")
	}
}

func TestGenerateExport_PDF_CreatesFile(t *testing.T) {
	reportRepo := newMockReportRepo()
	exportRepo := newMockExportRepo()
	exportPath := t.TempDir()
	svc := newReportService(reportRepo, exportRepo, exportPath)

	var reportID uuid.UUID
	for id := range reportRepo.reports {
		reportID = id
	}

	job, err := svc.GenerateExport(context.Background(), reportID, domain.ExportFormatPDF, map[string]string{"status": "active"}, uuid.New(), domain.UserRoleAdministrator)
	if err != nil {
		t.Fatalf("GenerateExport PDF failed: %v", err)
	}
	if job.Status != domain.ExportStatusCompleted {
		t.Errorf("expected status=completed, got %v", job.Status)
	}
	if job.FilePath == "" {
		t.Error("expected non-empty FilePath for completed PDF export")
	}
}

func TestGenerateExport_InvalidFormat_ErrValidation(t *testing.T) {
	reportRepo := newMockReportRepo()
	exportRepo := newMockExportRepo()
	svc := newReportService(reportRepo, exportRepo, t.TempDir())

	var reportID uuid.UUID
	for id := range reportRepo.reports {
		reportID = id
	}

	_, err := svc.GenerateExport(context.Background(), reportID, domain.ExportFormat("excel"), nil, uuid.New(), domain.UserRoleAdministrator)
	if err == nil {
		t.Fatal("expected ErrValidation for invalid export format, got nil")
	}
	var valErr *domain.ErrValidation
	if !isErrValidation(err, &valErr) {
		t.Errorf("expected *domain.ErrValidation, got %T: %v", err, err)
	}
}

// isErrValidation checks if err is of type *domain.ErrValidation using type assertion.
func isErrValidation(err error, target **domain.ErrValidation) bool {
	v, ok := err.(*domain.ErrValidation)
	if ok {
		*target = v
	}
	return ok
}

// newReportServiceImpl returns the concrete *ReportServiceImpl so tests can
// inject overrides (e.g. WithQueryDataFn) that are not on the interface.
func newReportServiceImpl(reportRepo *mockReportRepo, exportRepo *mockExportRepo, exportPath string) *application.ReportServiceImpl {
	cfg := platform.Config{ExportPath: exportPath}
	return application.NewReportService(reportRepo, exportRepo, cfg, &mockAuditService{}, nil)
}

func TestGenerateExport_QueryDataFailure_SetsStatusFailed(t *testing.T) {
	reportRepo := newMockReportRepo()
	exportRepo := newMockExportRepo()
	svc := newReportServiceImpl(reportRepo, exportRepo, t.TempDir())
	svc.WithQueryDataFn(func(_ context.Context, _ string, _ map[string]string) ([]map[string]interface{}, error) {
		return nil, errors.New("database unavailable")
	})

	var reportID uuid.UUID
	for id := range reportRepo.reports {
		reportID = id
	}

	job, err := svc.GenerateExport(context.Background(), reportID, domain.ExportFormatCSV, nil, uuid.New(), domain.UserRoleAdministrator)
	if err == nil {
		t.Fatal("expected error when queryReportData fails, got nil")
	}
	if job == nil {
		t.Fatal("expected job to be returned even on data query failure")
	}
	if job.Status != domain.ExportStatusFailed {
		t.Errorf("expected job.Status=failed, got %v", job.Status)
	}
}

func TestGenerateExport_QueryDataFailure_JobPersistedAsFailed(t *testing.T) {
	reportRepo := newMockReportRepo()
	exportRepo := newMockExportRepo()
	svc := newReportServiceImpl(reportRepo, exportRepo, t.TempDir())
	svc.WithQueryDataFn(func(_ context.Context, _ string, _ map[string]string) ([]map[string]interface{}, error) {
		return nil, errors.New("query error: connection reset")
	})

	var reportID uuid.UUID
	for id := range reportRepo.reports {
		reportID = id
	}

	job, err := svc.GenerateExport(context.Background(), reportID, domain.ExportFormatCSV, nil, uuid.New(), domain.UserRoleAdministrator)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if job == nil {
		t.Fatal("expected non-nil job on data query failure")
	}
	// Verify the repo was updated with the failed status (exportRepo.Update was called).
	persisted, repoErr := exportRepo.GetByID(context.Background(), job.ID)
	if repoErr != nil {
		t.Fatalf("job not found in export repo: %v", repoErr)
	}
	if persisted.Status != domain.ExportStatusFailed {
		t.Errorf("expected persisted job status=failed, got %v", persisted.Status)
	}
}
