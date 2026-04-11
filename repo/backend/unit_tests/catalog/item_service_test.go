package catalog_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
)

// --- Mock repositories ---

type mockItemRepo struct {
	items map[uuid.UUID]*domain.Item
}

func newMockItemRepo() *mockItemRepo {
	return &mockItemRepo{items: make(map[uuid.UUID]*domain.Item)}
}

func (m *mockItemRepo) Create(_ context.Context, item *domain.Item) error {
	m.items[item.ID] = item
	return nil
}
func (m *mockItemRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Item, error) {
	item, ok := m.items[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return item, nil
}
func (m *mockItemRepo) List(_ context.Context, _ map[string]string, _, _ int) ([]domain.Item, int, error) {
	list := make([]domain.Item, 0, len(m.items))
	for _, v := range m.items {
		list = append(list, *v)
	}
	return list, len(list), nil
}
func (m *mockItemRepo) Update(_ context.Context, item *domain.Item) error {
	if _, ok := m.items[item.ID]; !ok {
		return domain.ErrNotFound
	}
	m.items[item.ID] = item
	return nil
}
func (m *mockItemRepo) BatchUpdate(_ context.Context, items []*domain.Item) error {
	for _, item := range items {
		m.items[item.ID] = item
	}
	return nil
}

type mockAvailRepo struct {
	windows map[uuid.UUID][]domain.AvailabilityWindow
}

func newMockAvailRepo() *mockAvailRepo {
	return &mockAvailRepo{windows: make(map[uuid.UUID][]domain.AvailabilityWindow)}
}
func (m *mockAvailRepo) Create(_ context.Context, w *domain.AvailabilityWindow) error {
	m.windows[w.ItemID] = append(m.windows[w.ItemID], *w)
	return nil
}
func (m *mockAvailRepo) ListByItemID(_ context.Context, itemID uuid.UUID) ([]domain.AvailabilityWindow, error) {
	return m.windows[itemID], nil
}
func (m *mockAvailRepo) DeleteByItemID(_ context.Context, itemID uuid.UUID) error {
	delete(m.windows, itemID)
	return nil
}

type mockBlackoutRepo struct {
	windows map[uuid.UUID][]domain.BlackoutWindow
}

func newMockBlackoutRepo() *mockBlackoutRepo {
	return &mockBlackoutRepo{windows: make(map[uuid.UUID][]domain.BlackoutWindow)}
}
func (m *mockBlackoutRepo) Create(_ context.Context, w *domain.BlackoutWindow) error {
	m.windows[w.ItemID] = append(m.windows[w.ItemID], *w)
	return nil
}
func (m *mockBlackoutRepo) ListByItemID(_ context.Context, itemID uuid.UUID) ([]domain.BlackoutWindow, error) {
	return m.windows[itemID], nil
}
func (m *mockBlackoutRepo) DeleteByItemID(_ context.Context, itemID uuid.UUID) error {
	delete(m.windows, itemID)
	return nil
}

type mockBatchRepo struct {
	jobs    map[uuid.UUID]*domain.BatchEditJob
	results []domain.BatchEditResult
}

func newMockBatchRepo() *mockBatchRepo {
	return &mockBatchRepo{jobs: make(map[uuid.UUID]*domain.BatchEditJob)}
}
func (m *mockBatchRepo) CreateJob(_ context.Context, job *domain.BatchEditJob) error {
	m.jobs[job.ID] = job
	return nil
}
func (m *mockBatchRepo) CreateResult(_ context.Context, r *domain.BatchEditResult) error {
	m.results = append(m.results, *r)
	return nil
}
func (m *mockBatchRepo) GetJob(_ context.Context, id uuid.UUID) (*domain.BatchEditJob, error) {
	j, ok := m.jobs[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return j, nil
}
func (m *mockBatchRepo) ListResults(_ context.Context, _ uuid.UUID) ([]domain.BatchEditResult, error) {
	return m.results, nil
}

type mockAuditService struct{}

func (m *mockAuditService) Log(_ context.Context, _, _ string, _, _ uuid.UUID, _ map[string]interface{}) error {
	return nil
}
func (m *mockAuditService) List(_ context.Context, _ string, _ *uuid.UUID, _, _ int) ([]domain.AuditEvent, int, error) {
	return nil, 0, nil
}
func (m *mockAuditService) ListByEventTypes(_ context.Context, _ []string, _, _ int) ([]domain.AuditEvent, int, error) {
	return nil, 0, nil
}

// --- Tests ---

func TestItemCreate_ZeroDeposit_AppliesDefault(t *testing.T) {
	itemRepo := newMockItemRepo()
	availRepo := newMockAvailRepo()
	blackoutRepo := newMockBlackoutRepo()
	batchRepo := newMockBatchRepo()
	svc := application.NewItemService(itemRepo, availRepo, blackoutRepo, batchRepo, &mockAuditService{}, nil)

	item := &domain.Item{
		Name:              "Test Item",
		Category:          "Equipment",
		Brand:             "Acme",
		Condition:         domain.ItemConditionNew,
			BillingModel:      domain.BillingModelMonthlyRental,
		RefundableDeposit: 0, // should be defaulted
	}
	created, err := svc.Create(context.Background(), item, nil, nil)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.RefundableDeposit != domain.DefaultRefundableDeposit {
		t.Errorf("expected deposit=%v, got %v", domain.DefaultRefundableDeposit, created.RefundableDeposit)
	}
}

func TestItemCreate_MissingName_ReturnsValidationError(t *testing.T) {
	itemRepo := newMockItemRepo()
	availRepo := newMockAvailRepo()
	blackoutRepo := newMockBlackoutRepo()
	batchRepo := newMockBatchRepo()
	svc := application.NewItemService(itemRepo, availRepo, blackoutRepo, batchRepo, &mockAuditService{}, nil)

	item := &domain.Item{
		Category:     "Equipment",
		Brand:        "Acme",
		Condition:    domain.ItemConditionNew,
			BillingModel: domain.BillingModelMonthlyRental,
	}
	_, err := svc.Create(context.Background(), item, nil, nil)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestItemCreate_Success_StatusDraftVersion1(t *testing.T) {
	itemRepo := newMockItemRepo()
	availRepo := newMockAvailRepo()
	blackoutRepo := newMockBlackoutRepo()
	batchRepo := newMockBatchRepo()
	svc := application.NewItemService(itemRepo, availRepo, blackoutRepo, batchRepo, &mockAuditService{}, nil)

	item := &domain.Item{
		Name:         "Bike",
		Category:     "Equipment",
		Brand:        "Trek",
		Condition:    domain.ItemConditionNew,
			BillingModel: domain.BillingModelMonthlyRental,
	}
	created, err := svc.Create(context.Background(), item, nil, nil)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.Status != domain.ItemStatusDraft {
		t.Errorf("expected status=draft, got %v", created.Status)
	}
	if created.Version != 1 {
		t.Errorf("expected version=1, got %v", created.Version)
	}
}

func TestItemUpdate_VersionMismatch_ReturnsConflict(t *testing.T) {
	itemRepo := newMockItemRepo()
	availRepo := newMockAvailRepo()
	blackoutRepo := newMockBlackoutRepo()
	batchRepo := newMockBatchRepo()
	svc := application.NewItemService(itemRepo, availRepo, blackoutRepo, batchRepo, &mockAuditService{}, nil)

	existing := &domain.Item{
		ID:           uuid.New(),
		Name:         "Bike",
		Category:     "Equipment",
		Brand:        "Trek",
		Condition:    domain.ItemConditionNew,
			BillingModel: domain.BillingModelMonthlyRental,
		Status:       domain.ItemStatusDraft,
		Version:      2,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	itemRepo.items[existing.ID] = existing

	update := *existing
	update.Version = 1 // stale version
	_, err := svc.Update(context.Background(), &update, nil, nil)
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}
}

func TestItemPublish_PublishBlockedOnMissingFields(t *testing.T) {
	itemRepo := newMockItemRepo()
	availRepo := newMockAvailRepo()
	blackoutRepo := newMockBlackoutRepo()
	batchRepo := newMockBatchRepo()
	svc := application.NewItemService(itemRepo, availRepo, blackoutRepo, batchRepo, &mockAuditService{}, nil)

	item := &domain.Item{
		ID:           uuid.New(),
		Name:         "", // missing — should block publish
		Category:     "Equipment",
		Brand:        "Trek",
		Condition:    domain.ItemConditionNew,
			BillingModel: domain.BillingModelMonthlyRental,
		Status:       domain.ItemStatusDraft,
		Version:      1,
	}
	itemRepo.items[item.ID] = item

	err := svc.Publish(context.Background(), item.ID)
	var blocked *domain.ErrPublishBlocked
	if err == nil {
		t.Fatal("expected ErrPublishBlocked, got nil")
	}
	_ = blocked
	// Just check it fails — domain.ValidateItemForPublish catches missing name.
}

func TestItemPublish_ValidItem_StatusPublished(t *testing.T) {
	itemRepo := newMockItemRepo()
	availRepo := newMockAvailRepo()
	blackoutRepo := newMockBlackoutRepo()
	batchRepo := newMockBatchRepo()
	svc := application.NewItemService(itemRepo, availRepo, blackoutRepo, batchRepo, &mockAuditService{}, nil)

	item := &domain.Item{
		ID:           uuid.New(),
		Name:         "Bike",
		Category:     "Equipment",
		Brand:        "Trek",
		Condition:    domain.ItemConditionNew,
			BillingModel: domain.BillingModelMonthlyRental,
		Status:       domain.ItemStatusDraft,
		Version:      1,
		Quantity:     1,
	}
	itemRepo.items[item.ID] = item

	if err := svc.Publish(context.Background(), item.ID); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}
	if itemRepo.items[item.ID].Status != domain.ItemStatusPublished {
		t.Error("expected status=published")
	}
}

func TestItemUnpublish_DraftItem_ReturnsInvalidTransition(t *testing.T) {
	itemRepo := newMockItemRepo()
	availRepo := newMockAvailRepo()
	blackoutRepo := newMockBlackoutRepo()
	batchRepo := newMockBatchRepo()
	svc := application.NewItemService(itemRepo, availRepo, blackoutRepo, batchRepo, &mockAuditService{}, nil)

	item := &domain.Item{
		ID:           uuid.New(),
		Status:       domain.ItemStatusDraft,
		Name:         "Bike",
		Category:     "Equipment",
		Brand:        "Trek",
		Condition:    domain.ItemConditionNew,
			BillingModel: domain.BillingModelMonthlyRental,
		Version:      1,
	}
	itemRepo.items[item.ID] = item

	err := svc.Unpublish(context.Background(), item.ID)
	if err == nil {
		t.Fatal("expected invalid transition error, got nil")
	}
}

func TestItemUnpublish_PublishedItem_StatusDraft(t *testing.T) {
	itemRepo := newMockItemRepo()
	availRepo := newMockAvailRepo()
	blackoutRepo := newMockBlackoutRepo()
	batchRepo := newMockBatchRepo()
	svc := application.NewItemService(itemRepo, availRepo, blackoutRepo, batchRepo, &mockAuditService{}, nil)

	item := &domain.Item{
		ID:           uuid.New(),
		Status:       domain.ItemStatusPublished,
		Name:         "Bike",
		Category:     "Equipment",
		Brand:        "Trek",
		Condition:    domain.ItemConditionNew,
			BillingModel: domain.BillingModelMonthlyRental,
		Version:      2,
	}
	itemRepo.items[item.ID] = item

	if err := svc.Unpublish(context.Background(), item.ID); err != nil {
		t.Fatalf("Unpublish failed: %v", err)
	}
	if itemRepo.items[item.ID].Status != domain.ItemStatusDraft {
		t.Error("expected status=draft after unpublish")
	}
}

func TestBatchEdit_MixedResults(t *testing.T) {
	itemRepo := newMockItemRepo()
	availRepo := newMockAvailRepo()
	blackoutRepo := newMockBlackoutRepo()
	batchRepo := newMockBatchRepo()
	svc := application.NewItemService(itemRepo, availRepo, blackoutRepo, batchRepo, &mockAuditService{}, nil)

	existingID := uuid.New()
	itemRepo.items[existingID] = &domain.Item{
		ID:           existingID,
		Name:         "Bike",
		Category:     "Equipment",
		Brand:        "Trek",
		Condition:    domain.ItemConditionNew,
			BillingModel: domain.BillingModelMonthlyRental,
		Status:       domain.ItemStatusDraft,
		Version:      1,
		Quantity:     5,
	}
	createdBy := uuid.New()
	missingName := "x"
	unknownFieldValue := "x"
	newName := "New Name"

	missingID := uuid.New() // not in repo
	edits := []application.BatchEditInput{
		{ItemID: missingID, Field: "name", NewValue: &missingName},             // fail: not found
		{ItemID: existingID, Field: "unknown_field", NewValue: &unknownFieldValue}, // fail: invalid field
		{ItemID: existingID, Field: "name", NewValue: &newName},                // success
	}

	job, results, err := svc.BatchEdit(context.Background(), createdBy, edits)
	if err != nil {
		t.Fatalf("BatchEdit failed: %v", err)
	}
	if job.TotalRows != 3 {
		t.Errorf("expected TotalRows=3, got %d", job.TotalRows)
	}
	if job.SuccessCount != 1 {
		t.Errorf("expected SuccessCount=1, got %d", job.SuccessCount)
	}
	if job.FailureCount != 2 {
		t.Errorf("expected FailureCount=2, got %d", job.FailureCount)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Success {
		t.Error("expected first result to fail (item not found)")
	}
	if results[1].Success {
		t.Error("expected second result to fail (unknown field)")
	}
	if !results[2].Success {
		t.Error("expected third result to succeed")
	}
	// Verify the successful edit persisted.
	if itemRepo.items[existingID].Name != "New Name" {
		t.Errorf("expected name='New Name', got %q", itemRepo.items[existingID].Name)
	}
}
