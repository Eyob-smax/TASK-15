package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
)

// ReportStore

var _ store.ReportRepository = (*ReportStore)(nil)

type ReportStore struct {
	pool *pgxpool.Pool
}

func NewReportStore(pool *pgxpool.Pool) *ReportStore {
	return &ReportStore{pool: pool}
}

func (s *ReportStore) Create(ctx context.Context, r *domain.ReportDefinition) error {
	roleStrings := make([]string, len(r.AllowedRoles))
	for i, role := range r.AllowedRoles {
		roleStrings[i] = string(role)
	}

	filtersJSON, err := json.Marshal(r.Filters)
	if err != nil {
		return fmt.Errorf("report_store.Create marshal filters: %w", err)
	}

	_, err = s.pool.Exec(ctx,
		`INSERT INTO report_definitions
			(id, name, report_type, description, allowed_roles, filters, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		r.ID, r.Name, r.ReportType, r.Description, roleStrings, filtersJSON, r.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("report_store.Create: %w", err)
	}
	return nil
}

func (s *ReportStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.ReportDefinition, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, name, report_type, description, allowed_roles, filters, created_at
		 FROM report_definitions WHERE id=$1`,
		id,
	)
	r, err := scanReport(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("report_store.GetByID: %w", err)
	}
	return r, nil
}

func (s *ReportStore) List(ctx context.Context) ([]domain.ReportDefinition, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, report_type, description, allowed_roles, filters, created_at
		 FROM report_definitions ORDER BY name ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("report_store.List query: %w", err)
	}
	defer rows.Close()

	var reports []domain.ReportDefinition
	for rows.Next() {
		r, err := scanReport(rows)
		if err != nil {
			return nil, fmt.Errorf("report_store.List scan: %w", err)
		}
		reports = append(reports, *r)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("report_store.List rows: %w", err)
	}
	return reports, nil
}

func scanReport(row pgx.Row) (*domain.ReportDefinition, error) {
	var r domain.ReportDefinition
	var roleStrings []string
	var filtersJSON []byte

	err := row.Scan(
		&r.ID,
		&r.Name,
		&r.ReportType,
		&r.Description,
		&roleStrings,
		&filtersJSON,
		&r.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	r.AllowedRoles = make([]domain.UserRole, len(roleStrings))
	for i, s := range roleStrings {
		r.AllowedRoles[i] = domain.UserRole(s)
	}

	if err = json.Unmarshal(filtersJSON, &r.Filters); err != nil {
		return nil, fmt.Errorf("unmarshal filters: %w", err)
	}

	return &r, nil
}

// ExportStore

var _ store.ExportRepository = (*ExportStore)(nil)

type ExportStore struct {
	pool *pgxpool.Pool
}

func NewExportStore(pool *pgxpool.Pool) *ExportStore {
	return &ExportStore{pool: pool}
}

func (s *ExportStore) Create(ctx context.Context, job *domain.ExportJob) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO export_jobs
			(id, report_id, format, filename, status, file_path, created_by, created_at, completed_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		job.ID, job.ReportID, string(job.Format), job.Filename, string(job.Status),
		job.FilePath, job.CreatedBy, job.CreatedAt, job.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("export_store.Create: %w", err)
	}
	return nil
}

func (s *ExportStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.ExportJob, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, report_id, format, filename, status, file_path, created_by, created_at, completed_at
		 FROM export_jobs WHERE id=$1`,
		id,
	)
	job, err := scanExportJob(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("export_store.GetByID: %w", err)
	}
	return job, nil
}

func (s *ExportStore) Update(ctx context.Context, job *domain.ExportJob) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE export_jobs SET status=$1, file_path=$2, completed_at=$3 WHERE id=$4`,
		string(job.Status), job.FilePath, job.CompletedAt, job.ID,
	)
	if err != nil {
		return fmt.Errorf("export_store.Update: %w", err)
	}
	return nil
}

func scanExportJob(row pgx.Row) (*domain.ExportJob, error) {
	var job domain.ExportJob
	var format string
	var status string

	err := row.Scan(
		&job.ID,
		&job.ReportID,
		&format,
		&job.Filename,
		&status,
		&job.FilePath,
		&job.CreatedBy,
		&job.CreatedAt,
		&job.CompletedAt,
	)
	if err != nil {
		return nil, err
	}
	job.Format = domain.ExportFormat(format)
	job.Status = domain.ExportStatus(status)
	return &job, nil
}
