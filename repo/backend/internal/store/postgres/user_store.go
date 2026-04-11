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

// UserStore implements store.UserRepository using a pgx connection pool.
type UserStore struct {
	pool *pgxpool.Pool
}

// NewUserStore creates a new UserStore backed by the given connection pool.
func NewUserStore(pool *pgxpool.Pool) *UserStore {
	return &UserStore{pool: pool}
}

// GetByEmail retrieves a user by their email address. Returns domain.ErrNotFound
// if no matching user exists.
func (s *UserStore) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
		SELECT id, email, password_hash, salt, role, status, display_name,
		       location_id, failed_login_count, locked_until, created_at, updated_at
		FROM users
		WHERE email = $1
		LIMIT 1`

	row := s.pool.QueryRow(ctx, q, email)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("user_store.GetByEmail: %w", err)
	}
	return user, nil
}

// GetByID retrieves a user by their UUID. Returns domain.ErrNotFound if absent.
func (s *UserStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const q = `
		SELECT id, email, password_hash, salt, role, status, display_name,
		       location_id, failed_login_count, locked_until, created_at, updated_at
		FROM users
		WHERE id = $1
		LIMIT 1`

	row := s.pool.QueryRow(ctx, q, id)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("user_store.GetByID: %w", err)
	}
	return user, nil
}

// Update persists all mutable fields of a user record.
func (s *UserStore) Update(ctx context.Context, user *domain.User) error {
	const q = `
		UPDATE users
		SET email               = $1,
		    password_hash       = $2,
		    salt                = $3,
		    role                = $4,
		    status              = $5,
		    display_name        = $6,
		    location_id         = $7,
		    failed_login_count  = $8,
		    locked_until        = $9,
		    updated_at          = $10
		WHERE id = $11`

	_, err := s.pool.Exec(ctx, q,
		user.Email,
		user.PasswordHash,
		user.Salt,
		string(user.Role),
		string(user.Status),
		user.DisplayName,
		user.LocationID,
		user.FailedLoginCount,
		user.LockedUntil,
		time.Now().UTC(),
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("user_store.Update: %w", err)
	}
	return nil
}

// Create inserts a new user record into the database.
func (s *UserStore) Create(ctx context.Context, user *domain.User) error {
	const q = `
		INSERT INTO users
		    (id, email, password_hash, salt, role, status, display_name,
		     location_id, failed_login_count, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`

	_, err := s.pool.Exec(ctx, q,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Salt,
		string(user.Role),
		string(user.Status),
		user.DisplayName,
		user.LocationID,
		user.FailedLoginCount,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("user_store.Create: %w", err)
	}
	return nil
}

// List returns a paginated list of all users ordered by created_at DESC.
func (s *UserStore) List(ctx context.Context, page, pageSize int) ([]domain.User, int, error) {
	const countQ = `SELECT COUNT(*) FROM users`

	var total int
	if err := s.pool.QueryRow(ctx, countQ).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("user_store.List: %w", err)
	}

	const q = `
		SELECT id, email, password_hash, salt, role, status, display_name,
		       location_id, failed_login_count, locked_until, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	offset := (page - 1) * pageSize
	rows, err := s.pool.Query(ctx, q, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("user_store.List: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("user_store.List: %w", err)
		}
		users = append(users, *u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("user_store.List: %w", err)
	}
	return users, total, nil
}

// ListByRole returns a paginated list of users filtered by role.
func (s *UserStore) ListByRole(ctx context.Context, role domain.UserRole, page, pageSize int) ([]domain.User, int, error) {
	const countQ = `SELECT COUNT(*) FROM users WHERE role = $1`

	var total int
	if err := s.pool.QueryRow(ctx, countQ, string(role)).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("user_store.ListByRole: %w", err)
	}

	const q = `
		SELECT id, email, password_hash, salt, role, status, display_name,
		       location_id, failed_login_count, locked_until, created_at, updated_at
		FROM users
		WHERE role = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	offset := (page - 1) * pageSize
	rows, err := s.pool.Query(ctx, q, string(role), pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("user_store.ListByRole: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("user_store.ListByRole: %w", err)
		}
		users = append(users, *u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("user_store.ListByRole: %w", err)
	}
	return users, total, nil
}

// scanUser reads a single user row from a pgx row scanner.
func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	var roleStr, statusStr string
	var locationID *uuid.UUID

	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.Salt,
		&roleStr,
		&statusStr,
		&u.DisplayName,
		&locationID,
		&u.FailedLoginCount,
		&u.LockedUntil,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	u.Role = domain.UserRole(roleStr)
	u.Status = domain.UserStatus(statusStr)
	u.LocationID = locationID
	return &u, nil
}

// Compile-time interface assertion.
var _ store.UserRepository = (*UserStore)(nil)
