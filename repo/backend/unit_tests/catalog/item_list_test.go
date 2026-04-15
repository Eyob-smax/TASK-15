package catalog_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
)

func TestItemService_List_Paginated(t *testing.T) {
	itemRepo := newMockItemRepo()
	availRepo := newMockAvailRepo()
	blackoutRepo := newMockBlackoutRepo()
	batchRepo := newMockBatchRepo()
	svc := application.NewItemService(itemRepo, availRepo, blackoutRepo, batchRepo, &mockAuditService{}, nil)

	for i := 0; i < 3; i++ {
		id := uuid.New()
		itemRepo.items[id] = &domain.Item{
			ID:        id,
			Name:      "Item",
			Status:    domain.ItemStatusPublished,
			CreatedAt: time.Now(),
		}
	}

	rows, total, err := svc.List(context.Background(), 1, 10, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 3 || len(rows) != 3 {
		t.Errorf("expected 3, got total=%d rows=%d", total, len(rows))
	}
}

func TestItemService_List_WithFilters(t *testing.T) {
	itemRepo := newMockItemRepo()
	availRepo := newMockAvailRepo()
	blackoutRepo := newMockBlackoutRepo()
	batchRepo := newMockBatchRepo()
	svc := application.NewItemService(itemRepo, availRepo, blackoutRepo, batchRepo, &mockAuditService{}, nil)

	// Mock ignores filters, but we test that the call reaches the repo.
	_, _, err := svc.List(context.Background(), 1, 10, map[string]string{"status": "published"})
	if err != nil {
		t.Fatalf("List(filters) failed: %v", err)
	}
}
