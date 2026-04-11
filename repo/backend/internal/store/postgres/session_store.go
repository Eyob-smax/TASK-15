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

// SessionStore implements store.SessionRepository using a pgx connection pool.
type SessionStore struct {
	pool *pgxpool.Pool
}

// NewSessionStore creates a new SessionStore backed by the given connection pool.
func NewSessionStore(pool *pgxpool.Pool) *SessionStore {
	return &SessionStore{pool: pool}
}

// Create inserts a new session record into the database.
func (s *SessionStore) Create(ctx context.Context, session *domain.Session) error {
	const q = `
		INSERT INTO sessions (id, user_id, token, idle_expires_at, absolute_expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := s.pool.Exec(ctx, q,
		session.ID,
		session.UserID,
		session.Token,
		session.IdleExpiresAt,
		session.AbsoluteExpiresAt,
		session.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("session_store.Create: %w", err)
	}
	return nil
}

// GetByToken retrieves an active session by its token. Returns domain.ErrNotFound
// if the token does not exist.
func (s *SessionStore) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	const q = `
		SELECT id, user_id, token, idle_expires_at, absolute_expires_at, created_at
		FROM sessions
		WHERE token = $1
		LIMIT 1`

	row := s.pool.QueryRow(ctx, q, token)
	var sess domain.Session
	err := row.Scan(
		&sess.ID,
		&sess.UserID,
		&sess.Token,
		&sess.IdleExpiresAt,
		&sess.AbsoluteExpiresAt,
		&sess.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("session_store.GetByToken: %w", err)
	}
	return &sess, nil
}

// Delete removes a session by its UUID.
func (s *SessionStore) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM sessions WHERE id = $1`
	_, err := s.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("session_store.Delete: %w", err)
	}
	return nil
}

// DeleteExpired removes all sessions whose absolute_expires_at is before now.
// Returns the number of rows deleted.
func (s *SessionStore) DeleteExpired(ctx context.Context, now time.Time) (int, error) {
	const q = `DELETE FROM sessions WHERE absolute_expires_at < $1`
	tag, err := s.pool.Exec(ctx, q, now)
	if err != nil {
		return 0, fmt.Errorf("session_store.DeleteExpired: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

// UpdateIdleExpiry extends the idle_expires_at timestamp for a session. Called
// on every authenticated request to refresh the idle timeout.
func (s *SessionStore) UpdateIdleExpiry(ctx context.Context, id uuid.UUID, newExpiry time.Time) error {
	const q = `UPDATE sessions SET idle_expires_at = $1 WHERE id = $2`
	_, err := s.pool.Exec(ctx, q, newExpiry, id)
	if err != nil {
		return fmt.Errorf("session_store.UpdateIdleExpiry: %w", err)
	}
	return nil
}

// Compile-time interface assertion.
var _ store.SessionRepository = (*SessionStore)(nil)
