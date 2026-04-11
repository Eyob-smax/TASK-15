package application

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	gofpdf "github.com/jung-kurt/gofpdf"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/platform"
	"fitcommerce/internal/store"
)

// ReportServiceImpl implements ReportService.
type ReportServiceImpl struct {
	reportRepo  store.ReportRepository
	exportRepo  store.ExportRepository
	cfg         platform.Config
	auditSvc    AuditService
	pool        *pgxpool.Pool
	queryDataFn func(ctx context.Context, reportType string, filters map[string]string) ([]map[string]interface{}, error)
}

// WithQueryDataFn sets an override for queryReportData used in unit tests to
// inject failures without requiring a real database connection. Production code
// must never call this.
func (s *ReportServiceImpl) WithQueryDataFn(fn func(ctx context.Context, reportType string, filters map[string]string) ([]map[string]interface{}, error)) {
	s.queryDataFn = fn
}

// NewReportService creates a ReportServiceImpl backed by the given repositories.
func NewReportService(
	reportRepo store.ReportRepository,
	exportRepo store.ExportRepository,
	cfg platform.Config,
	auditSvc AuditService,
	pool *pgxpool.Pool,
) *ReportServiceImpl {
	return &ReportServiceImpl{
		reportRepo: reportRepo,
		exportRepo: exportRepo,
		cfg:        cfg,
		auditSvc:   auditSvc,
		pool:       pool,
	}
}

// List returns the report definitions accessible to the given user role.
func (s *ReportServiceImpl) List(ctx context.Context, userRole domain.UserRole) ([]domain.ReportDefinition, error) {
	reports, err := s.reportRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	filtered := make([]domain.ReportDefinition, 0, len(reports))
	for _, r := range reports {
		for _, allowed := range r.AllowedRoles {
			if allowed == userRole {
				filtered = append(filtered, r)
				break
			}
		}
	}

	return filtered, nil
}

// GetReport retrieves a report definition by ID.
func (s *ReportServiceImpl) GetReport(ctx context.Context, id uuid.UUID) (*domain.ReportDefinition, error) {
	return s.reportRepo.GetByID(ctx, id)
}

// GetData returns a structured dataset for the given report.
// callerRole is checked against the report's AllowedRoles as a second authorization layer.
func (s *ReportServiceImpl) GetData(ctx context.Context, reportID uuid.UUID, filters map[string]string, callerRole domain.UserRole) (interface{}, error) {
	report, err := s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return nil, err
	}

	allowed := false
	for _, r := range report.AllowedRoles {
		if r == callerRole {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	records, err := s.queryReportData(ctx, report.ReportType, filters)
	if err != nil {
		return nil, fmt.Errorf("report.GetData: %w", err)
	}

	return map[string]interface{}{
		"report_type": report.ReportType,
		"filters":     filters,
		"records":     records,
	}, nil
}

// queryReportData dispatches to the appropriate SQL query for each report type.
func (s *ReportServiceImpl) queryReportData(ctx context.Context, reportType string, filters map[string]string) ([]map[string]interface{}, error) {
	if s.queryDataFn != nil {
		return s.queryDataFn(ctx, reportType, filters)
	}
	if filters == nil {
		filters = map[string]string{}
	}
	switch reportType {
	case "member_growth":
		q, args := buildFilteredQuery(
			`SELECT id::text, membership_status, joined_at FROM members`,
			filters, map[string]string{"location_id": "location_id", "status": "membership_status", "_date": "joined_at"},
			" ORDER BY joined_at DESC LIMIT 500")
		return s.queryRows(ctx, q, args...)
	case "churn":
		q, args := buildFilteredQuery(
			`SELECT id::text, membership_status, updated_at FROM members WHERE membership_status IN ('expired','cancelled')`,
			filters, map[string]string{"location_id": "location_id", "_date": "updated_at"},
			" ORDER BY updated_at DESC LIMIT 500")
		return s.queryRows(ctx, q, args...)
	case "renewal_rate":
		q, args := buildFilteredQuery(
			`SELECT membership_status, COUNT(*)::int AS count FROM members`,
			filters, map[string]string{"location_id": "location_id"},
			" GROUP BY membership_status")
		return s.queryRows(ctx, q, args...)
	case "engagement":
		q, args := buildFilteredQuery(
			`SELECT id::text, status, quantity, total_amount, created_at FROM orders`,
			filters, map[string]string{"status": "status", "_date": "created_at"},
			" ORDER BY created_at DESC LIMIT 500")
		return s.queryRows(ctx, q, args...)
	case "class_fill_rate":
		q, args := buildFilteredQuery(
			`SELECT id::text, status, min_quantity, current_committed_qty, cutoff_time FROM group_buy_campaigns`,
			filters, map[string]string{"status": "status", "_date": "cutoff_time"},
			" ORDER BY cutoff_time DESC LIMIT 500")
		return s.queryRows(ctx, q, args...)
	case "coach_productivity":
		q, args := buildFilteredQuery(
			`SELECT c.id::text, c.specialization, c.is_active,
			        COUNT(m.id)::int AS member_count
			 FROM coaches c
			 LEFT JOIN members m ON m.location_id = c.location_id`,
			filters, map[string]string{"location_id": "c.location_id", "coach_id": "c.id"},
			" GROUP BY c.id, c.specialization, c.is_active")
		return s.queryRows(ctx, q, args...)
	case "inventory_summary":
		q, args := buildFilteredQuery(
			`SELECT id::text, name, category, quantity, status, unit_price FROM items`,
			filters, map[string]string{"category": "category", "status": "status"},
			" ORDER BY name LIMIT 500")
		return s.queryRows(ctx, q, args...)
	case "procurement_summary":
		q, args := buildFilteredQuery(
			`SELECT id::text, status, total_amount, created_at FROM purchase_orders`,
			filters, map[string]string{"status": "status", "_date": "created_at"},
			" ORDER BY created_at DESC LIMIT 500")
		return s.queryRows(ctx, q, args...)
	case "landed_cost_report":
		q, args := buildFilteredQuery(
			`SELECT item_id::text, purchase_order_id::text, period,
			        cost_component, raw_amount, allocated_amount, allocation_method
			 FROM landed_cost_entries`,
			filters, map[string]string{"item_id": "item_id", "purchase_order_id": "purchase_order_id", "period": "period", "_date": "created_at"},
			" ORDER BY created_at DESC LIMIT 500")
		return s.queryRows(ctx, q, args...)
	default:
		return []map[string]interface{}{}, nil
	}
}

// buildFilteredQuery appends parameterized WHERE clauses to a base query using
// the provided filter map and column mapping. The "_date" key in colMap defines
// which column is used for "from" and "to" range filters. Returns the final
// query string and the argument slice.
func buildFilteredQuery(baseQuery string, filters map[string]string, colMap map[string]string, suffix string) (string, []interface{}) {
	q := baseQuery
	args := []interface{}{}
	n := 1
	hasWhere := strings.Contains(strings.ToUpper(baseQuery), "WHERE")
	conj := func() string {
		if hasWhere {
			return " AND "
		}
		hasWhere = true
		return " WHERE "
	}

	for filterKey, colName := range colMap {
		if filterKey == "_date" {
			continue
		}
		if v, ok := filters[filterKey]; ok && v != "" {
			q += conj() + fmt.Sprintf("%s = $%d", colName, n)
			args = append(args, v)
			n++
		}
	}

	dateCol := colMap["_date"]
	if dateCol != "" {
		if v, ok := filters["from"]; ok && v != "" {
			q += conj() + fmt.Sprintf("%s >= $%d", dateCol, n)
			args = append(args, v)
			n++
		}
		if v, ok := filters["to"]; ok && v != "" {
			q += conj() + fmt.Sprintf("%s <= $%d", dateCol, n)
			args = append(args, v)
			n++
		}
	}

	q += suffix
	return q, args
}

// queryRows executes a query and returns each row as a map[string]interface{}.
func (s *ReportServiceImpl) queryRows(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fields := rows.FieldDescriptions()
	var result []map[string]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		row := make(map[string]interface{}, len(fields))
		for i, field := range fields {
			row[string(field.Name)] = values[i]
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, nil
}

// GenerateExport creates an export file for the given report in the requested
// format and records the job in the database.
// callerRole is checked against the report's AllowedRoles as a second authorization layer.
func (s *ReportServiceImpl) GenerateExport(ctx context.Context, reportID uuid.UUID, format domain.ExportFormat, parameters map[string]string, createdBy uuid.UUID, callerRole domain.UserRole) (*domain.ExportJob, error) {
	if !format.IsValid() {
		return nil, &domain.ErrValidation{Field: "format", Message: "unsupported export format"}
	}

	report, err := s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return nil, err
	}

	allowed := false
	for _, r := range report.AllowedRoles {
		if r == callerRole {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	filename := domain.GenerateExportFilename(report.ReportType, format, time.Now().UTC())

	job := &domain.ExportJob{
		ID:        uuid.New(),
		ReportID:  reportID,
		Format:    format,
		Filename:  filename,
		Status:    domain.ExportStatusPending,
		CreatedBy: createdBy,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.exportRepo.Create(ctx, job); err != nil {
		return nil, err
	}

	filePath := filepath.Join(s.cfg.ExportPath, filename)

	if err := os.MkdirAll(s.cfg.ExportPath, 0755); err != nil {
		job.Status = domain.ExportStatusFailed
		_ = s.exportRepo.Update(ctx, job)
		return job, err
	}

	records, dataErr := s.queryReportData(ctx, report.ReportType, sanitizeReportFilters(report.ReportType, parameters))
	if dataErr != nil {
		job.Status = domain.ExportStatusFailed
		_ = s.exportRepo.Update(ctx, job)
		return job, dataErr
	}
	// Apply role-based masking: only admins and operations managers see raw PII.
	records = maskExportRecords(records, callerRole)
	data := recordsToRows(report.ReportType, records)

	var writeErr error
	switch format {
	case domain.ExportFormatCSV:
		writeErr = writeCSV(filePath, data)
	case domain.ExportFormatPDF:
		writeErr = writePDF(filePath, report.ReportType, data)
	}

	if writeErr != nil {
		job.Status = domain.ExportStatusFailed
		_ = s.exportRepo.Update(ctx, job)
		return job, writeErr
	}

	now := time.Now().UTC()
	job.Status = domain.ExportStatusCompleted
	job.FilePath = filePath
	job.CompletedAt = &now

	if err := s.exportRepo.Update(ctx, job); err != nil {
		return job, err
	}

	_ = s.auditSvc.Log(ctx, "export.generated", "export_job", job.ID, createdBy, map[string]interface{}{
		"report_type": report.ReportType,
		"format":      string(format),
		"filename":    filename,
	})

	return job, nil
}

func sanitizeReportFilters(reportType string, filters map[string]string) map[string]string {
	if len(filters) == 0 {
		return nil
	}
	allowedByReportType := map[string]map[string]bool{
		"member_growth":      {"location_id": true, "status": true, "from": true, "to": true},
		"churn":              {"location_id": true, "from": true, "to": true},
		"renewal_rate":       {"location_id": true},
		"engagement":         {"status": true, "from": true, "to": true},
		"class_fill_rate":    {"status": true, "from": true, "to": true},
		"coach_productivity": {"location_id": true, "coach_id": true},
		"inventory_summary":  {"category": true, "status": true},
		"procurement_summary": {"status": true, "from": true, "to": true},
		"landed_cost_report": {"item_id": true, "purchase_order_id": true, "period": true, "from": true, "to": true},
	}

	allowed := allowedByReportType[reportType]
	if len(allowed) == 0 {
		return nil
	}

	sanitized := make(map[string]string)
	for key, value := range filters {
		if allowed[key] && value != "" {
			sanitized[key] = value
		}
	}
	return sanitized
}

// GetExport retrieves an export job by ID. Access is granted to the creator
// or to roles with administrative report access (administrator, operations_manager).
func (s *ReportServiceImpl) GetExport(ctx context.Context, id uuid.UUID, callerID uuid.UUID, callerRole domain.UserRole) (*domain.ExportJob, error) {
	job, err := s.exportRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	isOwner := job.CreatedBy == callerID
	isAdmin := callerRole == domain.UserRoleAdministrator || callerRole == domain.UserRoleOperationsManager
	if !isOwner && !isAdmin {
		return nil, domain.ErrForbidden
	}
	return job, nil
}

// DownloadExport returns the file path for a completed export job. Access is
// granted to the creator or to roles with administrative report access.
func (s *ReportServiceImpl) DownloadExport(ctx context.Context, id uuid.UUID, callerID uuid.UUID, callerRole domain.UserRole) (*domain.ExportJob, error) {
	job, err := s.exportRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	isOwner := job.CreatedBy == callerID
	isAdmin := callerRole == domain.UserRoleAdministrator || callerRole == domain.UserRoleOperationsManager
	if !isOwner && !isAdmin {
		return nil, domain.ErrForbidden
	}
	if job.Status != domain.ExportStatusCompleted {
		return nil, &domain.ErrConflict{Entity: "export", Message: "export is not ready for download"}
	}
	if job.FilePath == "" {
		return nil, domain.ErrNotFound
	}
	if _, err := os.Stat(job.FilePath); err != nil {
		if os.IsNotExist(err) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("report.DownloadExport stat: %w", err)
	}
	return job, nil
}

// recordsToRows converts report row maps to a tabular [][]string suitable for
// CSV/PDF export. The first row is a sorted header; subsequent rows are values.
// Falls back to a minimal metadata row when the records slice is empty.
func recordsToRows(reportType string, records []map[string]interface{}) [][]string {
	if len(records) == 0 {
		return [][]string{
			{"report_type", "generated_at"},
			{reportType, time.Now().UTC().Format(time.RFC3339)},
		}
	}
	headers := make([]string, 0, len(records[0]))
	for k := range records[0] {
		headers = append(headers, k)
	}
	sort.Strings(headers)

	table := make([][]string, 0, len(records)+1)
	table = append(table, headers)
	for _, rec := range records {
		row := make([]string, len(headers))
		for i, h := range headers {
			if v := rec[h]; v != nil {
				row[i] = fmt.Sprintf("%v", v)
			}
		}
		table = append(table, row)
	}
	return table
}

// maskExportRecords applies role-based field masking to export records.
// Administrators and OperationsManagers see all data unmasked.
// All other roles have email addresses masked and phone/contact fields redacted.
func maskExportRecords(records []map[string]interface{}, role domain.UserRole) []map[string]interface{} {
	if role == domain.UserRoleAdministrator || role == domain.UserRoleOperationsManager {
		return records
	}
	masked := make([]map[string]interface{}, len(records))
	for i, rec := range records {
		m := make(map[string]interface{}, len(rec))
		for k, v := range rec {
			switch k {
			case "email", "contact_email", "user_email":
				if s, ok := v.(string); ok {
					m[k] = maskExportEmail(s)
				} else {
					m[k] = v
				}
			case "phone", "contact_phone":
				m[k] = "[REDACTED]"
			default:
				m[k] = v
			}
		}
		masked[i] = m
	}
	return masked
}

// maskExportEmail masks an email address, revealing only the first character and domain.
func maskExportEmail(email string) string {
	idx := strings.Index(email, "@")
	if idx <= 0 {
		return "***@***"
	}
	return string(email[0]) + "***" + email[idx:]
}

// writeCSV writes the given records as a CSV file at path.
func writeCSV(path string, records [][]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.WriteAll(records); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

// writePDF writes the given records as a simple PDF file at path using gofpdf.
func writePDF(path string, title string, records [][]string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, title)
	pdf.Ln(12)
	pdf.SetFont("Arial", "", 10)
	for _, row := range records {
		for _, cell := range row {
			pdf.Cell(60, 8, cell)
		}
		pdf.Ln(8)
	}
	return pdf.OutputFileAndClose(path)
}

// Compile-time interface assertion.
var _ ReportService = (*ReportServiceImpl)(nil)
