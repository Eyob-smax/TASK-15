package procurement_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
)

// newVarianceService is a helper to build a VarianceService for tests.
func newVarianceService(varianceRepo *mockVarianceRepo, lineRepo *mockPOLineRepo, itemRepo *mockItemRepo, inventoryRepo *mockInventoryRepo) application.VarianceService {
	return application.NewVarianceService(varianceRepo, lineRepo, itemRepo, inventoryRepo, &mockAuditService{}, nil)
}

func TestVarianceGet_NotFound(t *testing.T) {
	varianceRepo := newMockVarianceRepo()
	svc := newVarianceService(varianceRepo, newMockPOLineRepo(), newMockItemRepo(), newMockInventoryRepo())

	_, err := svc.Get(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected ErrNotFound for missing variance, got nil")
	}
}

func TestVarianceGet_Found(t *testing.T) {
	varianceRepo := newMockVarianceRepo()
	svc := newVarianceService(varianceRepo, newMockPOLineRepo(), newMockItemRepo(), newMockInventoryRepo())

	id := uuid.New()
	record := &domain.VarianceRecord{
		ID:       id,
		POLineID: uuid.New(),
		Type:     domain.VarianceTypeShortage,
		Status:   domain.VarianceStatusOpen,
		CreatedAt: time.Now(),
	}
	varianceRepo.records[id] = record

	got, err := svc.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.ID != id {
		t.Errorf("expected ID=%v, got %v", id, got.ID)
	}
}

func TestVarianceList_OverdueAnnotation(t *testing.T) {
	varianceRepo := newMockVarianceRepo()

	yesterday := time.Now().Add(-25 * time.Hour)
	id := uuid.New()
	record := &domain.VarianceRecord{
		ID:                id,
		POLineID:          uuid.New(),
		Type:              domain.VarianceTypeShortage,
		Status:            domain.VarianceStatusOpen,
		ResolutionDueDate: yesterday,
		CreatedAt:         time.Now().Add(-48 * time.Hour),
	}
	varianceRepo.records[id] = record

	got := varianceRepo.records[id]
	if !got.IsOverdue(time.Now()) {
		t.Error("expected variance with past due date and Open status to be overdue")
	}
}

func TestVarianceList_NotOverdue(t *testing.T) {
	varianceRepo := newMockVarianceRepo()

	tomorrow := time.Now().Add(25 * time.Hour)
	id := uuid.New()
	record := &domain.VarianceRecord{
		ID:                id,
		POLineID:          uuid.New(),
		Type:              domain.VarianceTypeShortage,
		Status:            domain.VarianceStatusOpen,
		ResolutionDueDate: tomorrow,
		CreatedAt:         time.Now(),
	}
	varianceRepo.records[id] = record

	got := varianceRepo.records[id]
	if got.IsOverdue(time.Now()) {
		t.Error("expected variance with future due date to NOT be overdue")
	}
}

func TestVarianceResolve_Open_Success(t *testing.T) {
	varianceRepo := newMockVarianceRepo()
	lineRepo := newMockPOLineRepo()
	itemRepo := newMockItemRepo()
	inventoryRepo := newMockInventoryRepo()
	svc := newVarianceService(varianceRepo, lineRepo, itemRepo, inventoryRepo)

	id := uuid.New()
	lineID := uuid.New()
	itemID := uuid.New()
	record := &domain.VarianceRecord{
		ID:        id,
		POLineID:  lineID,
		Type:      domain.VarianceTypeShortage,
		Status:    domain.VarianceStatusOpen,
		CreatedAt: time.Now(),
	}
	varianceRepo.records[id] = record
	lineRepo.lines[lineID] = &domain.PurchaseOrderLine{ID: lineID, ItemID: itemID}
	itemRepo.items[itemID] = &domain.Item{ID: itemID, Quantity: 0, Version: 1}

	const notes = "resolved after supplier confirmation"
	change := 3
	if err := svc.Resolve(context.Background(), id, "adjustment", notes, &change, uuid.New()); err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	updated := varianceRepo.records[id]
	if updated.Status != domain.VarianceStatusResolved {
		t.Errorf("expected status=resolved, got %v", updated.Status)
	}
	if updated.ResolutionNotes != notes {
		t.Errorf("expected ResolutionNotes=%q, got %q", notes, updated.ResolutionNotes)
	}
	if updated.ResolutionAction != "adjustment" {
		t.Errorf("expected ResolutionAction=adjustment, got %q", updated.ResolutionAction)
	}
	if len(inventoryRepo.adjustments) != 1 || inventoryRepo.adjustments[0].QuantityChange != 3 {
		t.Error("expected inventory adjustment of +3")
	}
}

func TestVarianceResolve_Escalated_Success(t *testing.T) {
	varianceRepo := newMockVarianceRepo()
	lineRepo := newMockPOLineRepo()
	itemRepo := newMockItemRepo()
	inventoryRepo := newMockInventoryRepo()
	svc := newVarianceService(varianceRepo, lineRepo, itemRepo, inventoryRepo)

	id := uuid.New()
	lineID := uuid.New()
	itemID := uuid.New()
	record := &domain.VarianceRecord{
		ID:        id,
		POLineID:  lineID,
		Type:      domain.VarianceTypeShortage,
		Status:    domain.VarianceStatusEscalated,
		CreatedAt: time.Now(),
	}
	varianceRepo.records[id] = record
	lineRepo.lines[lineID] = &domain.PurchaseOrderLine{ID: lineID, ItemID: itemID}
	itemRepo.items[itemID] = &domain.Item{ID: itemID, Quantity: 0, Version: 1}

	change := 2
	if err := svc.Resolve(context.Background(), id, "adjustment", "resolved after escalation review", &change, uuid.New()); err != nil {
		t.Fatalf("Resolve failed for escalated variance: %v", err)
	}

	updated := varianceRepo.records[id]
	if updated.Status != domain.VarianceStatusResolved {
		t.Errorf("expected status=resolved, got %v", updated.Status)
	}
	if len(inventoryRepo.adjustments) != 1 || inventoryRepo.adjustments[0].QuantityChange != 2 {
		t.Error("expected inventory adjustment of +2 for escalated variance")
	}
}

func TestVarianceResolve_AlreadyResolved_ErrTransition(t *testing.T) {
	varianceRepo := newMockVarianceRepo()
	lineRepo := newMockPOLineRepo()
	itemRepo := newMockItemRepo()
	inventoryRepo := newMockInventoryRepo()
	svc := newVarianceService(varianceRepo, lineRepo, itemRepo, inventoryRepo)

	id := uuid.New()
	now := time.Now()
	lineID := uuid.New()
	record := &domain.VarianceRecord{
		ID:              id,
		POLineID:        lineID,
		Type:            domain.VarianceTypeShortage,
		Status:          domain.VarianceStatusResolved,
		ResolutionNotes: "already done",
		ResolvedAt:      &now,
		CreatedAt:       time.Now(),
	}
	varianceRepo.records[id] = record
	lineRepo.lines[lineID] = &domain.PurchaseOrderLine{ID: lineID, ItemID: uuid.New()}

	change := 1
	err := svc.Resolve(context.Background(), id, "adjustment", "try again", &change, uuid.New())
	if err == nil {
		t.Fatal("expected ErrInvalidTransition for already-resolved variance, got nil")
	}
}
