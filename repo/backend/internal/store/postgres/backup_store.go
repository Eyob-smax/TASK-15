package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
)

var _ store.BackupRepository = (*BackupStore)(nil)

type BackupStore struct {
	pool *pgxpool.Pool
}

func NewBackupStore(pool *pgxpool.Pool) *BackupStore {
	return &BackupStore{pool: pool}
}

func (s *BackupStore) Create(ctx context.Context, b *domain.BackupRun) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO backup_runs
			(id, archive_path, checksum, checksum_algorithm, encryption_key_ref,
			 status, file_size, started_at, completed_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		b.ID, b.ArchivePath, b.Checksum, b.ChecksumAlgorithm, b.EncryptionKeyRef,
		string(b.Status), b.FileSize, b.StartedAt, b.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("backup_store.Create: %w", err)
	}
	return nil
}

func (s *BackupStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.BackupRun, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, archive_path, checksum, checksum_algorithm, encryption_key_ref,
		        status, file_size, started_at, completed_at
		 FROM backup_runs WHERE id=$1`,
		id,
	)
	b, err := scanBackupRun(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("backup_store.GetByID: %w", err)
	}
	return b, nil
}

func (s *BackupStore) List(ctx context.Context, page, pageSize int) ([]domain.BackupRun, int, error) {
	offset := (page - 1) * pageSize

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM backup_runs`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("backup_store.List count: %w", err)
	}

	rows, err := s.pool.Query(ctx,
		`SELECT id, archive_path, checksum, checksum_algorithm, encryption_key_ref,
		        status, file_size, started_at, completed_at
		 FROM backup_runs ORDER BY started_at DESC LIMIT $1 OFFSET $2`,
		pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("backup_store.List query: %w", err)
	}
	defer rows.Close()

	var runs []domain.BackupRun
	for rows.Next() {
		b, err := scanBackupRun(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("backup_store.List scan: %w", err)
		}
		runs = append(runs, *b)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("backup_store.List rows: %w", err)
	}
	return runs, total, nil
}

func (s *BackupStore) Update(ctx context.Context, b *domain.BackupRun) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE backup_runs
		 SET archive_path=$1, checksum=$2, checksum_algorithm=$3, encryption_key_ref=$4,
		     status=$5, file_size=$6, completed_at=$7
		 WHERE id=$8`,
		b.ArchivePath, b.Checksum, b.ChecksumAlgorithm, b.EncryptionKeyRef,
		string(b.Status), b.FileSize, b.CompletedAt, b.ID,
	)
	if err != nil {
		return fmt.Errorf("backup_store.Update: %w", err)
	}
	return nil
}

func scanBackupRun(row pgx.Row) (*domain.BackupRun, error) {
	var b domain.BackupRun
	var status string

	err := row.Scan(
		&b.ID,
		&b.ArchivePath,
		&b.Checksum,
		&b.ChecksumAlgorithm,
		&b.EncryptionKeyRef,
		&status,
		&b.FileSize,
		&b.StartedAt,
		&b.CompletedAt,
	)
	if err != nil {
		return nil, err
	}
	b.Status = domain.BackupStatus(status)
	return &b, nil
}
