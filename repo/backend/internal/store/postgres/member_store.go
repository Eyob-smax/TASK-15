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

// MemberStore implements store.MemberRepository using a pgx connection pool.
type MemberStore struct {
	pool *pgxpool.Pool
}

// NewMemberStore creates a new MemberStore backed by the given pool.
func NewMemberStore(pool *pgxpool.Pool) *MemberStore {
	return &MemberStore{pool: pool}
}

// Create inserts a new member record.
func (s *MemberStore) Create(ctx context.Context, member *domain.Member) error {
	const q = `
		INSERT INTO members (id, user_id, location_id, membership_status, joined_at, renewal_date, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := s.pool.Exec(ctx, q,
		member.ID, member.UserID, member.LocationID, member.MembershipStatus,
		member.JoinedAt, member.RenewalDate, member.CreatedAt, member.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("member_store.Create: %w", err)
	}
	return nil
}

// GetByID retrieves a member by UUID.
func (s *MemberStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Member, error) {
	const q = `
		SELECT id, user_id, location_id, membership_status, joined_at, renewal_date, created_at, updated_at
		FROM members WHERE id = $1 LIMIT 1`
	row := s.pool.QueryRow(ctx, q, id)
	m, err := scanMember(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("member_store.GetByID: %w", err)
	}
	return m, nil
}

// GetByUserID retrieves the member record associated with a user.
func (s *MemberStore) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Member, error) {
	const q = `
		SELECT id, user_id, location_id, membership_status, joined_at, renewal_date, created_at, updated_at
		FROM members WHERE user_id = $1 LIMIT 1`
	row := s.pool.QueryRow(ctx, q, userID)
	m, err := scanMember(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("member_store.GetByUserID: %w", err)
	}
	return m, nil
}

// List returns a paginated list of members, optionally filtered by location.
func (s *MemberStore) List(ctx context.Context, locationID *uuid.UUID, page, pageSize int) ([]domain.Member, int, error) {
	offset := (page - 1) * pageSize

	var (
		total int
		rows  pgx.Rows
		err   error
	)

	if locationID != nil {
		if err = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM members WHERE location_id = $1`, *locationID).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("member_store.List count: %w", err)
		}
		rows, err = s.pool.Query(ctx, `
			SELECT id, user_id, location_id, membership_status, joined_at, renewal_date, created_at, updated_at
			FROM members WHERE location_id = $1 ORDER BY joined_at DESC LIMIT $2 OFFSET $3`,
			*locationID, pageSize, offset)
	} else {
		if err = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM members`).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("member_store.List count: %w", err)
		}
		rows, err = s.pool.Query(ctx, `
			SELECT id, user_id, location_id, membership_status, joined_at, renewal_date, created_at, updated_at
			FROM members ORDER BY joined_at DESC LIMIT $1 OFFSET $2`,
			pageSize, offset)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("member_store.List: %w", err)
	}
	defer rows.Close()

	var members []domain.Member
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("member_store.List scan: %w", err)
		}
		members = append(members, *m)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("member_store.List rows: %w", err)
	}
	return members, total, nil
}

// CountByPeriod returns the count of members who joined within the given time range.
func (s *MemberStore) CountByPeriod(ctx context.Context, locationID *uuid.UUID, start, end time.Time) (int, error) {
	var count int
	var err error
	if locationID != nil {
		err = s.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM members WHERE location_id = $1 AND joined_at >= $2 AND joined_at < $3`,
			*locationID, start, end).Scan(&count)
	} else {
		err = s.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM members WHERE joined_at >= $1 AND joined_at < $2`,
			start, end).Scan(&count)
	}
	if err != nil {
		return 0, fmt.Errorf("member_store.CountByPeriod: %w", err)
	}
	return count, nil
}

func scanMember(row pgx.Row) (*domain.Member, error) {
	var m domain.Member
	err := row.Scan(
		&m.ID, &m.UserID, &m.LocationID, &m.MembershipStatus,
		&m.JoinedAt, &m.RenewalDate, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// Compile-time interface assertion.
var _ store.MemberRepository = (*MemberStore)(nil)
