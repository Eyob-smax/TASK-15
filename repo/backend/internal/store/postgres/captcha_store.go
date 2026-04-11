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

// CaptchaStore implements store.CaptchaRepository using a pgx connection pool.
type CaptchaStore struct {
	pool *pgxpool.Pool
}

// NewCaptchaStore creates a new CaptchaStore backed by the given connection pool.
func NewCaptchaStore(pool *pgxpool.Pool) *CaptchaStore {
	return &CaptchaStore{pool: pool}
}

// Create inserts a new CAPTCHA challenge record using one-way verification
// material rather than persisting the plaintext answer.
func (s *CaptchaStore) Create(ctx context.Context, challenge *domain.CaptchaChallenge) error {
	const q = `
		INSERT INTO captcha_challenges (id, user_id, challenge_data, answer_hash, answer_salt, created_at, expires_at, verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := s.pool.Exec(ctx, q,
		challenge.ID,
		challenge.UserID,
		challenge.ChallengeData,
		challenge.AnswerHash,
		challenge.AnswerSalt,
		challenge.CreatedAt,
		challenge.ExpiresAt,
		challenge.Verified,
	)
	if err != nil {
		return fmt.Errorf("captcha_store.Create: %w", err)
	}
	return nil
}

// GetByID retrieves a CAPTCHA challenge by its UUID. Returns domain.ErrNotFound
// if the challenge does not exist.
func (s *CaptchaStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.CaptchaChallenge, error) {
	const q = `
		SELECT id, user_id, challenge_data, answer_hash, answer_salt, created_at, expires_at, verified
		FROM captcha_challenges
		WHERE id = $1
		LIMIT 1`

	row := s.pool.QueryRow(ctx, q, id)
	var c domain.CaptchaChallenge
	err := row.Scan(
		&c.ID,
		&c.UserID,
		&c.ChallengeData,
		&c.AnswerHash,
		&c.AnswerSalt,
		&c.CreatedAt,
		&c.ExpiresAt,
		&c.Verified,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("captcha_store.GetByID: %w", err)
	}
	return &c, nil
}

// MarkVerified sets the verified flag to true for the challenge identified by id.
func (s *CaptchaStore) MarkVerified(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE captcha_challenges SET verified = true WHERE id = $1`
	_, err := s.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("captcha_store.MarkVerified: %w", err)
	}
	return nil
}

// Compile-time interface assertion.
var _ store.CaptchaRepository = (*CaptchaStore)(nil)
