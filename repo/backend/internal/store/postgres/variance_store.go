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

// VarianceStore

var _ store.VarianceRepository = (*VarianceStore)(nil)

type VarianceStore struct {
	pool *pgxpool.Pool
}

func NewVarianceStore(pool *pgxpool.Pool) *VarianceStore {
	return &VarianceStore{pool: pool}
}

func (s *VarianceStore) Create(ctx context.Context, v *domain.VarianceRecord) error {
	db := executorFromContext(ctx, s.pool)
	_, err := db.Exec(ctx,
		`INSERT INTO variance_records
			(id, po_line_id, type, expected_value, actual_value, difference_amount,
			 status, resolution_due_date, resolved_at, resolution_action, resolution_notes, quantity_change, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		v.ID, v.POLineID, string(v.Type), v.ExpectedValue, v.ActualValue, v.DifferenceAmount,
		string(v.Status), v.ResolutionDueDate, v.ResolvedAt, v.ResolutionAction, v.ResolutionNotes, v.QuantityChange, v.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("variance_store.Create: %w", err)
	}
	return nil
}

func (s *VarianceStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.VarianceRecord, error) {
	db := executorFromContext(ctx, s.pool)
	row := db.QueryRow(ctx,
		`SELECT id, po_line_id, type, expected_value, actual_value, difference_amount,
		        status, resolution_due_date, resolved_at, resolution_action, resolution_notes, quantity_change, created_at
		 FROM variance_records WHERE id=$1`,
		id,
	)
	v, err := scanVariance(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("variance_store.GetByID: %w", err)
	}
	return v, nil
}

func (s *VarianceStore) List(ctx context.Context, status *domain.VarianceStatus, page, pageSize int) ([]domain.VarianceRecord, int, error) {
	db := executorFromContext(ctx, s.pool)
	offset := (page - 1) * pageSize

	var countRow pgx.Row
	var rows pgx.Rows
	var err error

	if status != nil {
		countRow = db.QueryRow(ctx,
			`SELECT COUNT(*) FROM variance_records WHERE status=$1`,
			string(*status),
		)
	} else {
		countRow = db.QueryRow(ctx, `SELECT COUNT(*) FROM variance_records`)
	}

	var total int
	if err = countRow.Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("variance_store.List count: %w", err)
	}

	if status != nil {
		rows, err = db.Query(ctx,
			`SELECT id, po_line_id, type, expected_value, actual_value, difference_amount,
			        status, resolution_due_date, resolved_at, resolution_action, resolution_notes, quantity_change, created_at
			 FROM variance_records WHERE status=$1
			 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
			string(*status), pageSize, offset,
		)
	} else {
		rows, err = db.Query(ctx,
			`SELECT id, po_line_id, type, expected_value, actual_value, difference_amount,
			        status, resolution_due_date, resolved_at, resolution_action, resolution_notes, quantity_change, created_at
			 FROM variance_records
			 ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			pageSize, offset,
		)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("variance_store.List query: %w", err)
	}
	defer rows.Close()

	var records []domain.VarianceRecord
	for rows.Next() {
		v, err := scanVariance(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("variance_store.List scan: %w", err)
		}
		records = append(records, *v)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("variance_store.List rows: %w", err)
	}
	return records, total, nil
}

func (s *VarianceStore) Update(ctx context.Context, v *domain.VarianceRecord) error {
	db := executorFromContext(ctx, s.pool)
	_, err := db.Exec(ctx,
		`UPDATE variance_records
		 SET status=$1, resolved_at=$2, resolution_action=$3, resolution_notes=$4, quantity_change=$5
		 WHERE id=$6`,
		string(v.Status), v.ResolvedAt, v.ResolutionAction, v.ResolutionNotes, v.QuantityChange, v.ID,
	)
	if err != nil {
		return fmt.Errorf("variance_store.Update: %w", err)
	}
	return nil
}

func (s *VarianceStore) ListUnresolved(ctx context.Context) ([]domain.VarianceRecord, error) {
	db := executorFromContext(ctx, s.pool)
	rows, err := db.Query(ctx,
		`SELECT id, po_line_id, type, expected_value, actual_value, difference_amount,
		        status, resolution_due_date, resolved_at, resolution_action, resolution_notes, quantity_change, created_at
		 FROM variance_records WHERE status='open'`,
	)
	if err != nil {
		return nil, fmt.Errorf("variance_store.ListUnresolved query: %w", err)
	}
	defer rows.Close()

	var records []domain.VarianceRecord
	for rows.Next() {
		v, err := scanVariance(rows)
		if err != nil {
			return nil, fmt.Errorf("variance_store.ListUnresolved scan: %w", err)
		}
		records = append(records, *v)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("variance_store.ListUnresolved rows: %w", err)
	}
	return records, nil
}

func scanVariance(row pgx.Row) (*domain.VarianceRecord, error) {
	var v domain.VarianceRecord
	var varType string
	var varStatus string
	var resolutionDueDate time.Time
	var quantityChange *int

	err := row.Scan(
		&v.ID,
		&v.POLineID,
		&varType,
		&v.ExpectedValue,
		&v.ActualValue,
		&v.DifferenceAmount,
		&varStatus,
		&resolutionDueDate,
		&v.ResolvedAt,
		&v.ResolutionAction,
		&v.ResolutionNotes,
		&quantityChange,
		&v.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	v.Type = domain.VarianceType(varType)
	v.Status = domain.VarianceStatus(varStatus)
	v.ResolutionDueDate = resolutionDueDate
	v.QuantityChange = quantityChange
	return &v, nil
}
