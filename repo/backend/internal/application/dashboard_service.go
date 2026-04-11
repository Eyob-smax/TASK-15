package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DashboardServiceImpl implements DashboardService using direct DB queries.
type DashboardServiceImpl struct {
	pool *pgxpool.Pool
}

// NewDashboardService creates a DashboardServiceImpl backed by the given pool.
func NewDashboardService(pool *pgxpool.Pool) *DashboardServiceImpl {
	return &DashboardServiceImpl{pool: pool}
}

// GetKPIs computes the six dashboard KPI metrics.
// locationID filters by location (nil = all). period determines the date range
// (daily/weekly/monthly/quarterly/yearly; default monthly). coachID filters
// member-related metrics by assigned coach (nil = all coaches). category
// filters engagement metrics by item category (empty = all). from/to override
// period bounds when both are provided (format: 2006-01-02).
func (s *DashboardServiceImpl) GetKPIs(ctx context.Context, locationID *uuid.UUID, period string, coachID *uuid.UUID, category string, from, to string) (*DashboardKPIs, error) {
	now := time.Now().UTC()
	curStart, curEnd, prevStart, prevEnd, periodLabel := computePeriodBounds(now, period, from, to)

	memberGrowth, err := s.memberGrowth(ctx, locationID, coachID, curStart, curEnd, prevStart, prevEnd, periodLabel)
	if err != nil {
		return nil, fmt.Errorf("dashboard.memberGrowth: %w", err)
	}

	churn, err := s.churn(ctx, locationID, coachID, curStart, curEnd, prevStart, prevEnd, periodLabel)
	if err != nil {
		return nil, fmt.Errorf("dashboard.churn: %w", err)
	}

	renewalRate, err := s.renewalRate(ctx, locationID, coachID, curStart, curEnd, prevStart, prevEnd, periodLabel)
	if err != nil {
		return nil, fmt.Errorf("dashboard.renewalRate: %w", err)
	}

	engagement, err := s.engagement(ctx, locationID, coachID, category, curStart, curEnd, prevStart, prevEnd, periodLabel)
	if err != nil {
		return nil, fmt.Errorf("dashboard.engagement: %w", err)
	}

	classFillRate, err := s.classFillRate(ctx, locationID, coachID, curStart, curEnd, prevStart, prevEnd, periodLabel)
	if err != nil {
		return nil, fmt.Errorf("dashboard.classFillRate: %w", err)
	}

	coachProductivity, err := s.coachProductivity(ctx, locationID, coachID, periodLabel)
	if err != nil {
		return nil, fmt.Errorf("dashboard.coachProductivity: %w", err)
	}

	return &DashboardKPIs{
		MemberGrowth:      memberGrowth,
		Churn:             churn,
		RenewalRate:       renewalRate,
		Engagement:        engagement,
		ClassFillRate:     classFillRate,
		CoachProductivity: coachProductivity,
	}, nil
}

// computePeriodBounds calculates current and previous period date ranges from the
// given period label. from/to override with a custom range when both are non-empty.
func computePeriodBounds(now time.Time, period, from, to string) (curStart, curEnd, prevStart, prevEnd time.Time, label string) {
	if from != "" && to != "" {
		curStart, _ = time.Parse("2006-01-02", from)
		curEnd, _ = time.Parse("2006-01-02", to)
		curStart = curStart.UTC()
		curEnd = curEnd.UTC().Add(24*time.Hour - time.Nanosecond)
		duration := curEnd.Sub(curStart)
		prevStart = curStart.Add(-duration)
		prevEnd = curStart
		label = from + "/" + to
		return
	}
	switch period {
	case "daily":
		curStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		curEnd = now
		prevStart = curStart.AddDate(0, 0, -1)
		prevEnd = curStart
		label = curStart.Format("2006-01-02")
	case "weekly":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		curStart = now.AddDate(0, 0, -(weekday - 1)).Truncate(24 * time.Hour)
		curEnd = now
		prevStart = curStart.AddDate(0, 0, -7)
		prevEnd = curStart
		_, week := curStart.ISOWeek()
		label = fmt.Sprintf("%d-W%02d", curStart.Year(), week)
	case "quarterly":
		quarter := (int(now.Month()) - 1) / 3
		curStart = time.Date(now.Year(), time.Month(quarter*3+1), 1, 0, 0, 0, 0, time.UTC)
		curEnd = now
		prevStart = curStart.AddDate(0, -3, 0)
		prevEnd = curStart
		label = fmt.Sprintf("%d-Q%d", now.Year(), quarter+1)
	case "yearly":
		curStart = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		curEnd = now
		prevStart = curStart.AddDate(-1, 0, 0)
		prevEnd = curStart
		label = fmt.Sprintf("%d", now.Year())
	default: // "monthly"
		curStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		curEnd = now
		prevStart = curStart.AddDate(0, -1, 0)
		prevEnd = curStart
		label = curStart.Format("2006-01")
	}
	return
}

func (s *DashboardServiceImpl) memberGrowth(ctx context.Context, locationID, coachID *uuid.UUID, curStart, curEnd, prevStart, prevEnd time.Time, period string) (KPIValue, error) {
	cur, err := s.countMembers(ctx, locationID, coachID, &curStart, &curEnd)
	if err != nil {
		return KPIValue{}, err
	}
	prev, err := s.countMembers(ctx, locationID, coachID, &prevStart, &prevEnd)
	if err != nil {
		return KPIValue{}, err
	}
	return makeKPI(float64(cur), float64(prev), period), nil
}

func (s *DashboardServiceImpl) churn(ctx context.Context, locationID, coachID *uuid.UUID, curStart, curEnd, prevStart, prevEnd time.Time, period string) (KPIValue, error) {
	cur, err := s.countChurnedMembers(ctx, locationID, coachID, &curStart, &curEnd)
	if err != nil {
		return KPIValue{}, err
	}
	prev, err := s.countChurnedMembers(ctx, locationID, coachID, &prevStart, &prevEnd)
	if err != nil {
		return KPIValue{}, err
	}
	return makeKPI(float64(cur), float64(prev), period), nil
}

func (s *DashboardServiceImpl) renewalRate(ctx context.Context, locationID, coachID *uuid.UUID, curStart, curEnd, prevStart, prevEnd time.Time, period string) (KPIValue, error) {
	curDue, err := s.countMembersDueForRenewal(ctx, locationID, coachID, &curStart, &curEnd)
	if err != nil {
		return KPIValue{}, err
	}
	curRenewed, err := s.countMembersRenewed(ctx, locationID, coachID, &curStart, &curEnd)
	if err != nil {
		return KPIValue{}, err
	}
	var curRate float64
	if curDue > 0 {
		curRate = float64(curRenewed) / float64(curDue) * 100
	}

	prevDue, err := s.countMembersDueForRenewal(ctx, locationID, coachID, &prevStart, &prevEnd)
	if err != nil {
		return KPIValue{}, err
	}
	prevRenewed, err := s.countMembersRenewed(ctx, locationID, coachID, &prevStart, &prevEnd)
	if err != nil {
		return KPIValue{}, err
	}
	var prevRate float64
	if prevDue > 0 {
		prevRate = float64(prevRenewed) / float64(prevDue) * 100
	}
	return makeKPI(curRate, prevRate, period), nil
}

func (s *DashboardServiceImpl) engagement(ctx context.Context, locationID, coachID *uuid.UUID, category string, curStart, curEnd, prevStart, prevEnd time.Time, period string) (KPIValue, error) {
	curOrders, err := s.countOrders(ctx, locationID, coachID, category, curStart, curEnd)
	if err != nil {
		return KPIValue{}, err
	}
	curMembers, err := s.countMembers(ctx, locationID, coachID, nil, nil)
	if err != nil {
		return KPIValue{}, err
	}
	prevOrders, err := s.countOrders(ctx, locationID, coachID, category, prevStart, prevEnd)
	if err != nil {
		return KPIValue{}, err
	}

	var curRate, prevRate float64
	if curMembers > 0 {
		curRate = float64(curOrders) / float64(curMembers)
		prevRate = float64(prevOrders) / float64(curMembers)
	}
	return makeKPI(curRate, prevRate, period), nil
}

func (s *DashboardServiceImpl) classFillRate(ctx context.Context, locationID, coachID *uuid.UUID, curStart, curEnd, prevStart, prevEnd time.Time, period string) (KPIValue, error) {
	curRate, err := s.avgClassCapacityOccupancy(ctx, locationID, coachID, &curStart, &curEnd)
	if err != nil {
		return KPIValue{}, err
	}
	prevRate, err := s.avgClassCapacityOccupancy(ctx, locationID, coachID, &prevStart, &prevEnd)
	if err != nil {
		return KPIValue{}, err
	}
	return makeKPI(curRate, prevRate, period), nil
}

func (s *DashboardServiceImpl) coachProductivity(ctx context.Context, locationID, coachID *uuid.UUID, period string) (KPIValue, error) {
	activeMembers, err := s.countMembersByStatus(ctx, locationID, coachID, "active")
	if err != nil {
		return KPIValue{}, err
	}
	activeCoaches, err := s.countActiveCoaches(ctx, locationID, coachID)
	if err != nil {
		return KPIValue{}, err
	}

	var ratio float64
	if activeCoaches > 0 {
		ratio = float64(activeMembers) / float64(activeCoaches)
	}
	// No meaningful previous period comparison without historical snapshots.
	return makeKPI(ratio, ratio, period), nil
}

// --- Helpers ---

func (s *DashboardServiceImpl) countMembers(ctx context.Context, locationID, coachID *uuid.UUID, start, end *time.Time) (int, error) {
	q := `SELECT COUNT(*) FROM members WHERE 1=1`
	args := []interface{}{}
	n := 1
	if locationID != nil {
		q += fmt.Sprintf(" AND location_id = $%d", n)
		args = append(args, *locationID)
		n++
	}
	if coachID != nil {
		q += fmt.Sprintf(" AND location_id = (SELECT location_id FROM coaches WHERE id = $%d)", n)
		args = append(args, *coachID)
		n++
	}
	if start != nil {
		q += fmt.Sprintf(" AND joined_at >= $%d", n)
		args = append(args, *start)
		n++
	}
	if end != nil {
		q += fmt.Sprintf(" AND joined_at < $%d", n)
		args = append(args, *end)
	}
	var count int
	if err := s.pool.QueryRow(ctx, q, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *DashboardServiceImpl) countMembersByStatus(ctx context.Context, locationID, coachID *uuid.UUID, status string) (int, error) {
	q := `SELECT COUNT(*) FROM members WHERE membership_status = $1`
	args := []interface{}{status}
	n := 2
	if locationID != nil {
		q += fmt.Sprintf(" AND location_id = $%d", n)
		args = append(args, *locationID)
		n++
	}
	if coachID != nil {
		q += fmt.Sprintf(" AND location_id = (SELECT location_id FROM coaches WHERE id = $%d)", n)
		args = append(args, *coachID)
	}
	var count int
	if err := s.pool.QueryRow(ctx, q, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *DashboardServiceImpl) countChurnedMembers(ctx context.Context, locationID, coachID *uuid.UUID, start, end *time.Time) (int, error) {
	q := `SELECT COUNT(*) FROM members WHERE membership_status IN ('expired','cancelled')`
	args := []interface{}{}
	n := 1
	if locationID != nil {
		q += fmt.Sprintf(" AND location_id = $%d", n)
		args = append(args, *locationID)
		n++
	}
	if coachID != nil {
		q += fmt.Sprintf(" AND location_id = (SELECT location_id FROM coaches WHERE id = $%d)", n)
		args = append(args, *coachID)
		n++
	}
	if start != nil {
		q += fmt.Sprintf(" AND updated_at >= $%d", n)
		args = append(args, *start)
		n++
	}
	if end != nil {
		q += fmt.Sprintf(" AND updated_at < $%d", n)
		args = append(args, *end)
	}
	var count int
	if err := s.pool.QueryRow(ctx, q, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *DashboardServiceImpl) countMembersDueForRenewal(ctx context.Context, locationID, coachID *uuid.UUID, start, end *time.Time) (int, error) {
	q := `SELECT COUNT(*) FROM members WHERE 1=1`
	args := []interface{}{}
	n := 1
	if locationID != nil {
		q += fmt.Sprintf(" AND location_id = $%d", n)
		args = append(args, *locationID)
		n++
	}
	if coachID != nil {
		q += fmt.Sprintf(" AND location_id = (SELECT location_id FROM coaches WHERE id = $%d)", n)
		args = append(args, *coachID)
		n++
	}
	if start != nil {
		q += fmt.Sprintf(" AND renewal_date >= $%d", n)
		args = append(args, *start)
		n++
	}
	if end != nil {
		q += fmt.Sprintf(" AND renewal_date < $%d", n)
		args = append(args, *end)
	}
	var count int
	if err := s.pool.QueryRow(ctx, q, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *DashboardServiceImpl) countMembersRenewed(ctx context.Context, locationID, coachID *uuid.UUID, start, end *time.Time) (int, error) {
	q := `SELECT COUNT(*) FROM members WHERE membership_status = 'active'`
	args := []interface{}{}
	n := 1
	if locationID != nil {
		q += fmt.Sprintf(" AND location_id = $%d", n)
		args = append(args, *locationID)
		n++
	}
	if coachID != nil {
		q += fmt.Sprintf(" AND location_id = (SELECT location_id FROM coaches WHERE id = $%d)", n)
		args = append(args, *coachID)
		n++
	}
	if start != nil {
		q += fmt.Sprintf(" AND renewal_date >= $%d", n)
		args = append(args, *start)
		n++
	}
	if end != nil {
		q += fmt.Sprintf(" AND renewal_date < $%d", n)
		args = append(args, *end)
	}
	var count int
	if err := s.pool.QueryRow(ctx, q, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *DashboardServiceImpl) countOrders(ctx context.Context, locationID, coachID *uuid.UUID, category string, start, end time.Time) (int, error) {
	q := `SELECT COUNT(*) FROM orders o JOIN items i ON i.id = o.item_id WHERE o.created_at >= $1 AND o.created_at < $2`
	args := []interface{}{start, end}
	n := 3

	if category != "" {
		q += fmt.Sprintf(" AND i.category = $%d", n)
		args = append(args, category)
		n++
	}
	if locationID != nil {
		q += fmt.Sprintf(" AND i.location_id = $%d", n)
		args = append(args, *locationID)
		n++
	}
	if coachID != nil {
		q += fmt.Sprintf(" AND i.location_id = (SELECT location_id FROM coaches WHERE id = $%d)", n)
		args = append(args, *coachID)
		n++
	}

	var count int
	if err := s.pool.QueryRow(ctx, q, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// avgClassCapacityOccupancy returns the average capacity occupancy across campaigns in
// the given period: AVG(current_committed_qty / max_quantity * 100) for campaigns where
// max_quantity (class capacity) is set. Campaigns without a capacity are excluded from
// the average. Returns 0 when no campaigns with capacity set exist in the period.
func (s *DashboardServiceImpl) avgClassCapacityOccupancy(ctx context.Context, locationID, coachID *uuid.UUID, start, end *time.Time) (float64, error) {
	q := `SELECT COALESCE(
	          AVG(CASE WHEN gbc.max_quantity IS NOT NULL AND gbc.max_quantity > 0
	              THEN gbc.current_committed_qty::float / gbc.max_quantity * 100
	              ELSE NULL END),
	          0)
	      FROM group_buy_campaigns gbc
	      JOIN items i ON i.id = gbc.item_id
	      WHERE 1=1`
	args := []interface{}{}
	n := 1
	if locationID != nil {
		q += fmt.Sprintf(" AND i.location_id = $%d", n)
		args = append(args, *locationID)
		n++
	}
	if coachID != nil {
		q += fmt.Sprintf(" AND i.location_id = (SELECT location_id FROM coaches WHERE id = $%d)", n)
		args = append(args, *coachID)
		n++
	}
	if start != nil {
		q += fmt.Sprintf(" AND gbc.created_at >= $%d", n)
		args = append(args, *start)
		n++
	}
	if end != nil {
		q += fmt.Sprintf(" AND gbc.created_at < $%d", n)
		args = append(args, *end)
	}
	var rate float64
	if err := s.pool.QueryRow(ctx, q, args...).Scan(&rate); err != nil {
		return 0, err
	}
	return rate, nil
}

func (s *DashboardServiceImpl) countActiveCoaches(ctx context.Context, locationID *uuid.UUID, coachID *uuid.UUID) (int, error) {
	q := `SELECT COUNT(*) FROM coaches WHERE is_active = true`
	args := []interface{}{}
	n := 1
	if locationID != nil {
		q += fmt.Sprintf(" AND location_id = $%d", n)
		args = append(args, *locationID)
		n++
	}
	if coachID != nil {
		q += fmt.Sprintf(" AND id = $%d", n)
		args = append(args, *coachID)
	}
	var count int
	if err := s.pool.QueryRow(ctx, q, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func makeKPI(value, previous float64, period string) KPIValue {
	var changePct float64
	if previous != 0 {
		changePct = (value - previous) / previous * 100
	}
	return KPIValue{
		Value:         value,
		PreviousValue: previous,
		ChangePercent: changePct,
		Period:        period,
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Compile-time interface assertion.
var _ DashboardService = (*DashboardServiceImpl)(nil)
