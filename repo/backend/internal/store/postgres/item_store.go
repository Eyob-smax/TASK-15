package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
)

// Compile-time interface compliance assertions.
var (
	_ store.ItemRepository               = (*ItemStore)(nil)
	_ store.BatchEditRepository          = (*ItemStore)(nil)
	_ store.AvailabilityWindowRepository = (*AvailabilityWindowStore)(nil)
	_ store.BlackoutWindowRepository     = (*BlackoutWindowStore)(nil)
)

// ItemStore implements ItemRepository and BatchEditRepository.
type ItemStore struct {
	pool *pgxpool.Pool
}

// NewItemStore creates a new ItemStore backed by the given connection pool.
func NewItemStore(pool *pgxpool.Pool) *ItemStore {
	return &ItemStore{pool: pool}
}

// AvailabilityWindowStore implements AvailabilityWindowRepository.
type AvailabilityWindowStore struct {
	pool *pgxpool.Pool
}

// NewAvailabilityWindowStore creates a new AvailabilityWindowStore.
func NewAvailabilityWindowStore(pool *pgxpool.Pool) *AvailabilityWindowStore {
	return &AvailabilityWindowStore{pool: pool}
}

// BlackoutWindowStore implements BlackoutWindowRepository.
type BlackoutWindowStore struct {
	pool *pgxpool.Pool
}

// NewBlackoutWindowStore creates a new BlackoutWindowStore.
func NewBlackoutWindowStore(pool *pgxpool.Pool) *BlackoutWindowStore {
	return &BlackoutWindowStore{pool: pool}
}

// --- ItemRepository ---

func (s *ItemStore) Create(ctx context.Context, item *domain.Item) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		INSERT INTO items
			(id, sku, name, description, category, brand, condition, unit_price,
			 refundable_deposit, billing_model, status, quantity, location_id,
			 created_by, created_at, updated_at, version)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`
	_, err := db.Exec(ctx, q,
		item.ID, item.SKU, item.Name, item.Description, item.Category, item.Brand,
		string(item.Condition), item.UnitPrice, item.RefundableDeposit,
		string(item.BillingModel), string(item.Status), item.Quantity, item.LocationID,
		item.CreatedBy, item.CreatedAt, item.UpdatedAt, item.Version,
	)
	return err
}

func (s *ItemStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Item, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `
		SELECT id, sku, name, description, category, brand, condition, unit_price,
		       refundable_deposit, billing_model, status, quantity, location_id,
		       created_by, created_at, updated_at, version
		FROM items WHERE id = $1`
	row := db.QueryRow(ctx, q, id)
	item, err := scanItem(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return item, nil
}

func (s *ItemStore) List(ctx context.Context, filters map[string]string, page, pageSize int) ([]domain.Item, int, error) {
	db := executorFromContext(ctx, s.pool)
	conditions := []string{}
	args := []interface{}{}
	idx := 1

	for _, col := range []string{"category", "brand", "condition", "status"} {
		if v, ok := filters[col]; ok && v != "" {
			conditions = append(conditions, fmt.Sprintf("%s = $%d", col, idx))
			args = append(args, v)
			idx++
		}
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM items %s", where)
	var total int
	if err := db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery := fmt.Sprintf(`
		SELECT id, sku, name, description, category, brand, condition, unit_price,
		       refundable_deposit, billing_model, status, quantity, location_id,
		       created_by, created_at, updated_at, version
		FROM items %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, where, idx, idx+1)
	listArgs := append(args, pageSize, (page-1)*pageSize)

	rows, err := db.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []domain.Item
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *item)
	}
	return items, total, rows.Err()
}

// Update applies a conditional update using optimistic concurrency. The query
// matches on both id and version-1 (the version before increment). If zero
// rows are affected the record was modified concurrently and ErrConflict is
// returned.
func (s *ItemStore) Update(ctx context.Context, item *domain.Item) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		UPDATE items SET
			sku=$1, name=$2, description=$3, category=$4, brand=$5, condition=$6,
			unit_price=$7, refundable_deposit=$8, billing_model=$9, status=$10,
			quantity=$11, location_id=$12, updated_at=$13, version=$14
		WHERE id=$15 AND version=$16`
	tag, err := db.Exec(ctx, q,
		item.SKU, item.Name, item.Description, item.Category, item.Brand,
		string(item.Condition), item.UnitPrice, item.RefundableDeposit,
		string(item.BillingModel), string(item.Status), item.Quantity, item.LocationID,
		item.UpdatedAt, item.Version, item.ID, item.Version-1,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return &domain.ErrConflict{Entity: "item", Message: "concurrent modification detected"}
	}
	return nil
}

func (s *ItemStore) BatchUpdate(ctx context.Context, items []*domain.Item) error {
	batch := &pgx.Batch{}
	const q = `
		UPDATE items SET
			sku=$1, name=$2, description=$3, category=$4, brand=$5, condition=$6,
			unit_price=$7, refundable_deposit=$8, billing_model=$9, status=$10,
			quantity=$11, location_id=$12, updated_at=$13, version=$14
		WHERE id=$15 AND version=$16`
	for _, item := range items {
		batch.Queue(q,
			item.SKU, item.Name, item.Description, item.Category, item.Brand,
			string(item.Condition), item.UnitPrice, item.RefundableDeposit,
			string(item.BillingModel), string(item.Status), item.Quantity, item.LocationID,
			item.UpdatedAt, item.Version, item.ID, item.Version-1,
		)
	}
	results := s.pool.SendBatch(ctx, batch)
	defer results.Close()
	for range items {
		if _, err := results.Exec(); err != nil {
			return err
		}
	}
	return nil
}

// --- AvailabilityWindowRepository ---

func (s *AvailabilityWindowStore) Create(ctx context.Context, window *domain.AvailabilityWindow) error {
	db := executorFromContext(ctx, s.pool)
	const q = `INSERT INTO item_availability_windows (id, item_id, start_time, end_time) VALUES ($1,$2,$3,$4)`
	_, err := db.Exec(ctx, q, window.ID, window.ItemID, window.StartTime, window.EndTime)
	return err
}

func (s *AvailabilityWindowStore) ListByItemID(ctx context.Context, itemID uuid.UUID) ([]domain.AvailabilityWindow, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `SELECT id, item_id, start_time, end_time FROM item_availability_windows WHERE item_id = $1 ORDER BY start_time`
	rows, err := db.Query(ctx, q, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var windows []domain.AvailabilityWindow
	for rows.Next() {
		var w domain.AvailabilityWindow
		if err := rows.Scan(&w.ID, &w.ItemID, &w.StartTime, &w.EndTime); err != nil {
			return nil, err
		}
		windows = append(windows, w)
	}
	return windows, rows.Err()
}

func (s *AvailabilityWindowStore) DeleteByItemID(ctx context.Context, itemID uuid.UUID) error {
	db := executorFromContext(ctx, s.pool)
	_, err := db.Exec(ctx, `DELETE FROM item_availability_windows WHERE item_id = $1`, itemID)
	return err
}

// --- BlackoutWindowRepository ---

func (s *BlackoutWindowStore) Create(ctx context.Context, window *domain.BlackoutWindow) error {
	db := executorFromContext(ctx, s.pool)
	const q = `INSERT INTO item_blackout_windows (id, item_id, start_time, end_time) VALUES ($1,$2,$3,$4)`
	_, err := db.Exec(ctx, q, window.ID, window.ItemID, window.StartTime, window.EndTime)
	return err
}

func (s *BlackoutWindowStore) ListByItemID(ctx context.Context, itemID uuid.UUID) ([]domain.BlackoutWindow, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `SELECT id, item_id, start_time, end_time FROM item_blackout_windows WHERE item_id = $1 ORDER BY start_time`
	rows, err := db.Query(ctx, q, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var windows []domain.BlackoutWindow
	for rows.Next() {
		var w domain.BlackoutWindow
		if err := rows.Scan(&w.ID, &w.ItemID, &w.StartTime, &w.EndTime); err != nil {
			return nil, err
		}
		windows = append(windows, w)
	}
	return windows, rows.Err()
}

func (s *BlackoutWindowStore) DeleteByItemID(ctx context.Context, itemID uuid.UUID) error {
	db := executorFromContext(ctx, s.pool)
	_, err := db.Exec(ctx, `DELETE FROM item_blackout_windows WHERE item_id = $1`, itemID)
	return err
}

// --- BatchEditRepository (on ItemStore) ---

func (s *ItemStore) CreateJob(ctx context.Context, job *domain.BatchEditJob) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		INSERT INTO batch_edit_jobs (id, created_by, created_at, total_rows, success_count, failure_count)
		VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := db.Exec(ctx, q,
		job.ID, job.CreatedBy, job.CreatedAt,
		job.TotalRows, job.SuccessCount, job.FailureCount,
	)
	return err
}

func (s *ItemStore) CreateResult(ctx context.Context, result *domain.BatchEditResult) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		INSERT INTO batch_edit_results
			(id, batch_id, item_id, field, old_value, new_value, success, failure_reason)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := db.Exec(ctx, q,
		result.ID, result.BatchID, result.ItemID,
		result.Field, result.OldValue, result.NewValue,
		result.Success, result.FailureReason,
	)
	return err
}

func (s *ItemStore) GetJob(ctx context.Context, id uuid.UUID) (*domain.BatchEditJob, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `
		SELECT id, created_by, created_at, total_rows, success_count, failure_count
		FROM batch_edit_jobs WHERE id = $1`
	var job domain.BatchEditJob
	err := db.QueryRow(ctx, q, id).Scan(
		&job.ID, &job.CreatedBy, &job.CreatedAt,
		&job.TotalRows, &job.SuccessCount, &job.FailureCount,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &job, nil
}

func (s *ItemStore) ListResults(ctx context.Context, batchID uuid.UUID) ([]domain.BatchEditResult, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `
		SELECT id, batch_id, item_id, field, old_value, new_value, success, failure_reason
		FROM batch_edit_results WHERE batch_id = $1 ORDER BY id`
	rows, err := db.Query(ctx, q, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.BatchEditResult
	for rows.Next() {
		var r domain.BatchEditResult
		if err := rows.Scan(&r.ID, &r.BatchID, &r.ItemID, &r.Field,
			&r.OldValue, &r.NewValue, &r.Success, &r.FailureReason); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// --- Private helpers ---

type scannable interface {
	Scan(dest ...interface{}) error
}

func scanItem(row scannable) (*domain.Item, error) {
	var item domain.Item
	var condition, billingModel, status string
	var updatedAt time.Time
	err := row.Scan(
		&item.ID, &item.SKU, &item.Name, &item.Description, &item.Category, &item.Brand,
		&condition, &item.UnitPrice, &item.RefundableDeposit, &billingModel,
		&status, &item.Quantity, &item.LocationID,
		&item.CreatedBy, &item.CreatedAt, &updatedAt, &item.Version,
	)
	if err != nil {
		return nil, err
	}
	item.Condition = domain.ItemCondition(condition)
	item.BillingModel = domain.BillingModel(billingModel)
	item.Status = domain.ItemStatus(status)
	item.UpdatedAt = updatedAt
	return &item, nil
}
