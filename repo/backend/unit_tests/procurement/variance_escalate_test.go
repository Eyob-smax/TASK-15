package procurement_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
)

func TestVariance_EscalateOverdue_EscalatesOverdueOnly(t *testing.T) {
	varianceRepo := newMockVarianceRepo()
	svc := newVarianceService(varianceRepo, newMockPOLineRepo(), newMockItemRepo(), newMockInventoryRepo())

	past := time.Now().Add(-48 * time.Hour)
	future := time.Now().Add(48 * time.Hour)

	overdueID := uuid.New()
	futureID := uuid.New()
	resolvedID := uuid.New()

	varianceRepo.records[overdueID] = &domain.VarianceRecord{
		ID:                overdueID,
		POLineID:          uuid.New(),
		Type:              domain.VarianceTypeShortage,
		Status:            domain.VarianceStatusOpen,
		ResolutionDueDate: past,
	}
	varianceRepo.records[futureID] = &domain.VarianceRecord{
		ID:                futureID,
		POLineID:          uuid.New(),
		Type:              domain.VarianceTypeShortage,
		Status:            domain.VarianceStatusOpen,
		ResolutionDueDate: future,
	}
	varianceRepo.records[resolvedID] = &domain.VarianceRecord{
		ID:                resolvedID,
		POLineID:          uuid.New(),
		Type:              domain.VarianceTypeShortage,
		Status:            domain.VarianceStatusResolved,
		ResolutionDueDate: past,
	}

	count, err := svc.EscalateOverdue(context.Background())
	if err != nil {
		t.Fatalf("EscalateOverdue failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 record escalated, got %d", count)
	}
	if varianceRepo.records[overdueID].Status != domain.VarianceStatusEscalated {
		t.Errorf("overdue record should be escalated, got %v", varianceRepo.records[overdueID].Status)
	}
	if varianceRepo.records[futureID].Status != domain.VarianceStatusOpen {
		t.Errorf("future record should remain open, got %v", varianceRepo.records[futureID].Status)
	}
	if varianceRepo.records[resolvedID].Status != domain.VarianceStatusResolved {
		t.Errorf("resolved record should stay resolved, got %v", varianceRepo.records[resolvedID].Status)
	}
}

func TestVariance_EscalateOverdue_NoRecords(t *testing.T) {
	varianceRepo := newMockVarianceRepo()
	svc := newVarianceService(varianceRepo, newMockPOLineRepo(), newMockItemRepo(), newMockInventoryRepo())

	count, err := svc.EscalateOverdue(context.Background())
	if err != nil {
		t.Fatalf("EscalateOverdue failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 escalated, got %d", count)
	}
}
