package procurement_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
)

type landedCostRepoStub struct {
	entries      []domain.LandedCostEntry
	byItemPeriod []domain.LandedCostEntry
	byPO         []domain.LandedCostEntry
	itemErr      error
	poErr        error
}

func (m *landedCostRepoStub) Create(_ context.Context, e *domain.LandedCostEntry) error {
	m.entries = append(m.entries, *e)
	return nil
}

func (m *landedCostRepoStub) ListByItemAndPeriod(_ context.Context, _ uuid.UUID, _ string) ([]domain.LandedCostEntry, error) {
	if m.itemErr != nil {
		return nil, m.itemErr
	}
	return m.byItemPeriod, nil
}

func (m *landedCostRepoStub) ListByPOID(_ context.Context, _ uuid.UUID) ([]domain.LandedCostEntry, error) {
	if m.poErr != nil {
		return nil, m.poErr
	}
	return m.byPO, nil
}

func TestLandedCost_GetSummary_ReturnsEntries(t *testing.T) {
	repo := &landedCostRepoStub{byItemPeriod: []domain.LandedCostEntry{
		{ID: uuid.New(), ItemID: uuid.New(), Period: "2026-Q1", AllocatedAmount: 10.0},
	}}
	svc := application.NewLandedCostService(repo)

	entries, err := svc.GetSummary(context.Background(), uuid.New(), "2026-Q1")
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestLandedCost_GetSummary_RepoError(t *testing.T) {
	repo := &landedCostRepoStub{itemErr: errors.New("db err")}
	svc := application.NewLandedCostService(repo)

	_, err := svc.GetSummary(context.Background(), uuid.New(), "2026-Q1")
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestLandedCost_GetByPOID_ReturnsEntries(t *testing.T) {
	repo := &landedCostRepoStub{byPO: []domain.LandedCostEntry{
		{ID: uuid.New(), PurchaseOrderID: uuid.New(), AllocatedAmount: 50.0},
		{ID: uuid.New(), PurchaseOrderID: uuid.New(), AllocatedAmount: 25.0},
	}}
	svc := application.NewLandedCostService(repo)

	entries, err := svc.GetByPOID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("GetByPOID failed: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestLandedCost_GetByPOID_RepoError(t *testing.T) {
	repo := &landedCostRepoStub{poErr: errors.New("db err")}
	svc := application.NewLandedCostService(repo)

	_, err := svc.GetByPOID(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
