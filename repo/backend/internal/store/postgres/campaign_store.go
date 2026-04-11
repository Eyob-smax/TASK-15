package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
)

// Compile-time interface compliance assertions.
var (
	_ store.CampaignRepository    = (*CampaignStore)(nil)
	_ store.ParticipantRepository = (*ParticipantStore)(nil)
)

// CampaignStore implements CampaignRepository.
type CampaignStore struct {
	pool *pgxpool.Pool
}

// NewCampaignStore creates a new CampaignStore backed by the given pool.
func NewCampaignStore(pool *pgxpool.Pool) *CampaignStore {
	return &CampaignStore{pool: pool}
}

// ParticipantStore implements ParticipantRepository.
type ParticipantStore struct {
	pool *pgxpool.Pool
}

// NewParticipantStore creates a new ParticipantStore backed by the given pool.
func NewParticipantStore(pool *pgxpool.Pool) *ParticipantStore {
	return &ParticipantStore{pool: pool}
}

// --- CampaignRepository ---

func (s *CampaignStore) Create(ctx context.Context, campaign *domain.GroupBuyCampaign) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		INSERT INTO group_buy_campaigns
			(id, item_id, min_quantity, max_quantity, current_committed_qty, cutoff_time,
			 status, created_by, created_at, evaluated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := db.Exec(ctx, q,
		campaign.ID, campaign.ItemID, campaign.MinQuantity, campaign.MaxQuantity,
		campaign.CurrentCommittedQty, campaign.CutoffTime,
		string(campaign.Status), campaign.CreatedBy, campaign.CreatedAt,
		campaign.EvaluatedAt,
	)
	return err
}

func (s *CampaignStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.GroupBuyCampaign, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `
		SELECT id, item_id, min_quantity, max_quantity, current_committed_qty, cutoff_time,
		       status, created_by, created_at, evaluated_at
		FROM group_buy_campaigns WHERE id = $1`
	campaign, err := scanCampaign(db.QueryRow(ctx, q, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return campaign, nil
}

func (s *CampaignStore) List(ctx context.Context, page, pageSize int) ([]domain.GroupBuyCampaign, int, error) {
	db := executorFromContext(ctx, s.pool)
	var total int
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM group_buy_campaigns`).Scan(&total); err != nil {
		return nil, 0, err
	}

	const q = `
		SELECT id, item_id, min_quantity, max_quantity, current_committed_qty, cutoff_time,
		       status, created_by, created_at, evaluated_at
		FROM group_buy_campaigns
		ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := db.Query(ctx, q, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var campaigns []domain.GroupBuyCampaign
	for rows.Next() {
		c, err := scanCampaign(rows)
		if err != nil {
			return nil, 0, err
		}
		campaigns = append(campaigns, *c)
	}
	return campaigns, total, rows.Err()
}

func (s *CampaignStore) Update(ctx context.Context, campaign *domain.GroupBuyCampaign) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		UPDATE group_buy_campaigns SET
			current_committed_qty=$1, status=$2, evaluated_at=$3
		WHERE id=$4`
	_, err := db.Exec(ctx, q,
		campaign.CurrentCommittedQty,
		string(campaign.Status),
		campaign.EvaluatedAt,
		campaign.ID,
	)
	return err
}

// ListActive returns campaigns with status 'active' whose cutoff is in the future.
func (s *CampaignStore) ListActive(ctx context.Context) ([]domain.GroupBuyCampaign, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `
		SELECT id, item_id, min_quantity, max_quantity, current_committed_qty, cutoff_time,
		       status, created_by, created_at, evaluated_at
		FROM group_buy_campaigns
		WHERE status = 'active' AND cutoff_time > NOW()
		ORDER BY cutoff_time`
	rows, err := db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []domain.GroupBuyCampaign
	for rows.Next() {
		c, err := scanCampaign(rows)
		if err != nil {
			return nil, err
		}
		campaigns = append(campaigns, *c)
	}
	return campaigns, rows.Err()
}

// ListDueCampaigns returns active campaigns whose cutoff_at is at or before now.
// Used by CutoffEvalJob to find campaigns that require evaluation.
func (s *CampaignStore) ListDueCampaigns(ctx context.Context, now time.Time) ([]domain.GroupBuyCampaign, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `
		SELECT id, item_id, min_quantity, max_quantity, current_committed_qty, cutoff_time,
		       status, created_by, created_at, evaluated_at
		FROM group_buy_campaigns
		WHERE status = 'active' AND cutoff_time <= $1
		ORDER BY cutoff_time`
	rows, err := db.Query(ctx, q, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []domain.GroupBuyCampaign
	for rows.Next() {
		c, err := scanCampaign(rows)
		if err != nil {
			return nil, err
		}
		campaigns = append(campaigns, *c)
	}
	return campaigns, rows.Err()
}

// --- ParticipantRepository ---

func (s *ParticipantStore) Create(ctx context.Context, participant *domain.GroupBuyParticipant) error {
	db := executorFromContext(ctx, s.pool)
	const q = `
		INSERT INTO group_buy_participants (id, campaign_id, user_id, quantity, order_id, joined_at)
		VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := db.Exec(ctx, q,
		participant.ID, participant.CampaignID, participant.UserID,
		participant.Quantity, participant.OrderID, participant.JoinedAt,
	)
	return err
}

func (s *ParticipantStore) ListByCampaign(ctx context.Context, campaignID uuid.UUID) ([]domain.GroupBuyParticipant, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `
		SELECT id, campaign_id, user_id, quantity, order_id, joined_at
		FROM group_buy_participants WHERE campaign_id = $1 ORDER BY joined_at`
	rows, err := db.Query(ctx, q, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []domain.GroupBuyParticipant
	for rows.Next() {
		var p domain.GroupBuyParticipant
		if err := rows.Scan(&p.ID, &p.CampaignID, &p.UserID, &p.Quantity, &p.OrderID, &p.JoinedAt); err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}
	return participants, rows.Err()
}

func (s *ParticipantStore) CountCommittedQuantity(ctx context.Context, campaignID uuid.UUID) (int, error) {
	db := executorFromContext(ctx, s.pool)
	const q = `
		SELECT COALESCE(SUM(p.quantity), 0)
		FROM group_buy_participants p
		JOIN orders o ON o.id = p.order_id
		WHERE p.campaign_id = $1
		  AND o.status IN ('created', 'paid')`
	var total int
	err := db.QueryRow(ctx, q, campaignID).Scan(&total)
	return total, err
}

// --- Private helpers ---

type campaignScannable interface {
	Scan(dest ...interface{}) error
}

func scanCampaign(row campaignScannable) (*domain.GroupBuyCampaign, error) {
	var c domain.GroupBuyCampaign
	var status string
	err := row.Scan(
		&c.ID, &c.ItemID, &c.MinQuantity, &c.MaxQuantity, &c.CurrentCommittedQty,
		&c.CutoffTime, &status, &c.CreatedBy, &c.CreatedAt, &c.EvaluatedAt,
	)
	if err != nil {
		return nil, err
	}
	c.Status = domain.CampaignStatus(status)
	return &c, nil
}
