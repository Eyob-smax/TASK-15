package procurement_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
)

// --- Mock repositories ---

type mockPORepo struct {
	orders map[uuid.UUID]*domain.PurchaseOrder
}

func newMockPORepo() *mockPORepo {
	return &mockPORepo{orders: make(map[uuid.UUID]*domain.PurchaseOrder)}
}

func (m *mockPORepo) Create(_ context.Context, po *domain.PurchaseOrder) error {
	m.orders[po.ID] = po
	return nil
}

func (m *mockPORepo) GetByID(_ context.Context, id uuid.UUID) (*domain.PurchaseOrder, error) {
	po, ok := m.orders[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return po, nil
}

func (m *mockPORepo) List(_ context.Context, _, _ int) ([]domain.PurchaseOrder, int, error) {
	list := make([]domain.PurchaseOrder, 0, len(m.orders))
	for _, v := range m.orders {
		list = append(list, *v)
	}
	return list, len(list), nil
}

func (m *mockPORepo) Update(_ context.Context, po *domain.PurchaseOrder) error {
	if _, ok := m.orders[po.ID]; !ok {
		return domain.ErrNotFound
	}
	m.orders[po.ID] = po
	return nil
}

type mockPOLineRepo struct {
	lines map[uuid.UUID]*domain.PurchaseOrderLine // keyed by line ID
	byPO  map[uuid.UUID][]uuid.UUID               // poID -> line IDs
}

func newMockPOLineRepo() *mockPOLineRepo {
	return &mockPOLineRepo{
		lines: make(map[uuid.UUID]*domain.PurchaseOrderLine),
		byPO:  make(map[uuid.UUID][]uuid.UUID),
	}
}

func (m *mockPOLineRepo) Create(_ context.Context, line *domain.PurchaseOrderLine) error {
	m.lines[line.ID] = line
	m.byPO[line.PurchaseOrderID] = append(m.byPO[line.PurchaseOrderID], line.ID)
	return nil
}

func (m *mockPOLineRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.PurchaseOrderLine, error) {
	line, ok := m.lines[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return line, nil
}

func (m *mockPOLineRepo) ListByPOID(_ context.Context, poID uuid.UUID) ([]domain.PurchaseOrderLine, error) {
	ids := m.byPO[poID]
	result := make([]domain.PurchaseOrderLine, 0, len(ids))
	for _, id := range ids {
		result = append(result, *m.lines[id])
	}
	return result, nil
}

func (m *mockPOLineRepo) Update(_ context.Context, line *domain.PurchaseOrderLine) error {
	if _, ok := m.lines[line.ID]; !ok {
		return domain.ErrNotFound
	}
	m.lines[line.ID] = line
	return nil
}

type mockVarianceRepo struct {
	records map[uuid.UUID]*domain.VarianceRecord
}

func newMockVarianceRepo() *mockVarianceRepo {
	return &mockVarianceRepo{records: make(map[uuid.UUID]*domain.VarianceRecord)}
}

func (m *mockVarianceRepo) Create(_ context.Context, r *domain.VarianceRecord) error {
	m.records[r.ID] = r
	return nil
}

func (m *mockVarianceRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.VarianceRecord, error) {
	r, ok := m.records[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return r, nil
}

func (m *mockVarianceRepo) List(_ context.Context, _ *domain.VarianceStatus, _, _ int) ([]domain.VarianceRecord, int, error) {
	list := make([]domain.VarianceRecord, 0, len(m.records))
	for _, v := range m.records {
		list = append(list, *v)
	}
	return list, len(list), nil
}

func (m *mockVarianceRepo) Update(_ context.Context, r *domain.VarianceRecord) error {
	if _, ok := m.records[r.ID]; !ok {
		return domain.ErrNotFound
	}
	m.records[r.ID] = r
	return nil
}

func (m *mockVarianceRepo) ListUnresolved(_ context.Context) ([]domain.VarianceRecord, error) {
	var list []domain.VarianceRecord
	for _, v := range m.records {
		if v.Status == domain.VarianceStatusOpen {
			list = append(list, *v)
		}
	}
	return list, nil
}

type mockLandedCostRepo struct {
	entries []domain.LandedCostEntry
}

func newMockLandedCostRepo() *mockLandedCostRepo {
	return &mockLandedCostRepo{}
}

func (m *mockLandedCostRepo) Create(_ context.Context, e *domain.LandedCostEntry) error {
	m.entries = append(m.entries, *e)
	return nil
}

func (m *mockLandedCostRepo) ListByItemAndPeriod(_ context.Context, _ uuid.UUID, _ string) ([]domain.LandedCostEntry, error) {
	return m.entries, nil
}

func (m *mockLandedCostRepo) ListByPOID(_ context.Context, _ uuid.UUID) ([]domain.LandedCostEntry, error) {
	return m.entries, nil
}

type mockInventoryRepo struct {
	adjustments []domain.InventoryAdjustment
	snapshots   []domain.InventorySnapshot
}

func newMockInventoryRepo() *mockInventoryRepo {
	return &mockInventoryRepo{}
}

func (m *mockInventoryRepo) CreateSnapshot(_ context.Context, snapshot *domain.InventorySnapshot) error {
	m.snapshots = append(m.snapshots, *snapshot)
	return nil
}

func (m *mockInventoryRepo) ListSnapshots(_ context.Context, _ *uuid.UUID, _ *uuid.UUID) ([]domain.InventorySnapshot, error) {
	return nil, nil
}

func (m *mockInventoryRepo) CreateAdjustment(_ context.Context, adj *domain.InventoryAdjustment) error {
	m.adjustments = append(m.adjustments, *adj)
	return nil
}

func (m *mockInventoryRepo) ListAdjustments(_ context.Context, _ *uuid.UUID, _, _ int) ([]domain.InventoryAdjustment, int, error) {
	return m.adjustments, len(m.adjustments), nil
}

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

// --- Helper to build a PurchaseOrderService ---

func newPOService(poRepo *mockPORepo, lineRepo *mockPOLineRepo, varianceRepo *mockVarianceRepo, landedCostRepo *mockLandedCostRepo, inventoryRepo *mockInventoryRepo, itemRepo *mockItemRepo) application.PurchaseOrderService {
	return application.NewPurchaseOrderService(poRepo, lineRepo, varianceRepo, landedCostRepo, inventoryRepo, itemRepo, &mockAuditService{}, nil)
}

// --- Tests ---

func TestCreate_ItemNotFound(t *testing.T) {
	poRepo := newMockPORepo()
	lineRepo := newMockPOLineRepo()
	varianceRepo := newMockVarianceRepo()
	landedCostRepo := newMockLandedCostRepo()
	inventoryRepo := newMockInventoryRepo()
	itemRepo := newMockItemRepo()

	svc := newPOService(poRepo, lineRepo, varianceRepo, landedCostRepo, inventoryRepo, itemRepo)

	po := &domain.PurchaseOrder{
		SupplierID: uuid.New(),
		CreatedBy:  uuid.New(),
	}
	lines := []domain.PurchaseOrderLine{
		{ItemID: uuid.New(), OrderedQuantity: 5, OrderedUnitPrice: 10.00},
	}

	_, err := svc.Create(context.Background(), po, lines)
	if err == nil {
		t.Fatal("expected ErrNotFound for missing item, got nil")
	}
}

func TestCreate_Success(t *testing.T) {
	poRepo := newMockPORepo()
	lineRepo := newMockPOLineRepo()
	varianceRepo := newMockVarianceRepo()
	landedCostRepo := newMockLandedCostRepo()
	inventoryRepo := newMockInventoryRepo()
	itemRepo := newMockItemRepo()

	// Pre-seed items so the service can validate them.
	item1ID := uuid.New()
	item2ID := uuid.New()
	itemRepo.items[item1ID] = &domain.Item{ID: item1ID, Name: "Item 1"}
	itemRepo.items[item2ID] = &domain.Item{ID: item2ID, Name: "Item 2"}

	svc := newPOService(poRepo, lineRepo, varianceRepo, landedCostRepo, inventoryRepo, itemRepo)

	po := &domain.PurchaseOrder{
		SupplierID: uuid.New(),
		CreatedBy:  uuid.New(),
	}
	lines := []domain.PurchaseOrderLine{
		{ItemID: item1ID, OrderedQuantity: 3, OrderedUnitPrice: 20.00},
		{ItemID: item2ID, OrderedQuantity: 2, OrderedUnitPrice: 15.00},
	}

	created, err := svc.Create(context.Background(), po, lines)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.Status != domain.POStatusCreated {
		t.Errorf("expected status=created, got %v", created.Status)
	}
	// total = 3*20 + 2*15 = 60 + 30 = 90
	if created.TotalAmount != 90.00 {
		t.Errorf("expected TotalAmount=90.00, got %v", created.TotalAmount)
	}
	// lineRepo should have 2 lines for this PO.
	storedLines, err := lineRepo.ListByPOID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("ListByPOID failed: %v", err)
	}
	if len(storedLines) != 2 {
		t.Errorf("expected 2 lines in lineRepo, got %d", len(storedLines))
	}
}

func TestApprove_CreatedToApproved(t *testing.T) {
	poRepo := newMockPORepo()
	lineRepo := newMockPOLineRepo()
	varianceRepo := newMockVarianceRepo()
	landedCostRepo := newMockLandedCostRepo()
	inventoryRepo := newMockInventoryRepo()
	itemRepo := newMockItemRepo()

	svc := newPOService(poRepo, lineRepo, varianceRepo, landedCostRepo, inventoryRepo, itemRepo)

	poID := uuid.New()
	approverID := uuid.New()
	po := &domain.PurchaseOrder{
		ID:        poID,
		Status:    domain.POStatusCreated,
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		Version:   1,
	}
	poRepo.orders[poID] = po

	if err := svc.Approve(context.Background(), poID, approverID); err != nil {
		t.Fatalf("Approve failed: %v", err)
	}

	updated := poRepo.orders[poID]
	if updated.Status != domain.POStatusApproved {
		t.Errorf("expected status=approved, got %v", updated.Status)
	}
	if updated.ApprovedBy == nil || *updated.ApprovedBy != approverID {
		t.Error("expected ApprovedBy to be set to approverID")
	}
}

func TestApprove_AlreadyReceived_ErrTransition(t *testing.T) {
	poRepo := newMockPORepo()
	lineRepo := newMockPOLineRepo()
	varianceRepo := newMockVarianceRepo()
	landedCostRepo := newMockLandedCostRepo()
	inventoryRepo := newMockInventoryRepo()
	itemRepo := newMockItemRepo()

	svc := newPOService(poRepo, lineRepo, varianceRepo, landedCostRepo, inventoryRepo, itemRepo)

	poID := uuid.New()
	po := &domain.PurchaseOrder{
		ID:        poID,
		Status:    domain.POStatusReceived,
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		Version:   1,
	}
	poRepo.orders[poID] = po

	err := svc.Approve(context.Background(), poID, uuid.New())
	if err == nil {
		t.Fatal("expected ErrInvalidTransition for Received->Approve, got nil")
	}
}

func TestReceive_QtyMismatch_CreatesVariance(t *testing.T) {
	poRepo := newMockPORepo()
	lineRepo := newMockPOLineRepo()
	varianceRepo := newMockVarianceRepo()
	landedCostRepo := newMockLandedCostRepo()
	inventoryRepo := newMockInventoryRepo()
	itemRepo := newMockItemRepo()

	svc := newPOService(poRepo, lineRepo, varianceRepo, landedCostRepo, inventoryRepo, itemRepo)

	poID := uuid.New()
	po := &domain.PurchaseOrder{
		ID:        poID,
		Status:    domain.POStatusApproved,
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		Version:   1,
	}
	poRepo.orders[poID] = po

	lineID := uuid.New()
	line := &domain.PurchaseOrderLine{
		ID:               lineID,
		PurchaseOrderID:  poID,
		ItemID:           uuid.New(),
		OrderedQuantity:  10,
		OrderedUnitPrice: 5.00,
	}
	itemRepo.items[line.ItemID] = &domain.Item{ID: line.ItemID, Quantity: 0, Version: 1}
	lineRepo.lines[lineID] = line
	lineRepo.byPO[poID] = append(lineRepo.byPO[poID], lineID)

	// Receive with different quantity (8 vs 10).
	err := svc.Receive(context.Background(), poID, []application.ReceivedLineInput{
		{POLineID: lineID, ReceivedQuantity: 8, ReceivedUnitPrice: 5.00},
	}, uuid.New())
	if err != nil {
		t.Fatalf("Receive failed: %v", err)
	}

	if len(varianceRepo.records) != 1 {
		t.Errorf("expected 1 variance record (qty mismatch), got %d", len(varianceRepo.records))
	}
}

func TestReceive_PriceMismatch_CreatesVariance(t *testing.T) {
	poRepo := newMockPORepo()
	lineRepo := newMockPOLineRepo()
	varianceRepo := newMockVarianceRepo()
	landedCostRepo := newMockLandedCostRepo()
	inventoryRepo := newMockInventoryRepo()
	itemRepo := newMockItemRepo()

	svc := newPOService(poRepo, lineRepo, varianceRepo, landedCostRepo, inventoryRepo, itemRepo)

	poID := uuid.New()
	po := &domain.PurchaseOrder{
		ID:        poID,
		Status:    domain.POStatusApproved,
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		Version:   1,
	}
	poRepo.orders[poID] = po

	lineID := uuid.New()
	line := &domain.PurchaseOrderLine{
		ID:               lineID,
		PurchaseOrderID:  poID,
		ItemID:           uuid.New(),
		OrderedQuantity:  10,
		OrderedUnitPrice: 5.00,
	}
	itemRepo.items[line.ItemID] = &domain.Item{ID: line.ItemID, Quantity: 0, Version: 1}
	lineRepo.lines[lineID] = line
	lineRepo.byPO[poID] = append(lineRepo.byPO[poID], lineID)

	// Receive with different unit price (6.00 vs 5.00), same qty.
	err := svc.Receive(context.Background(), poID, []application.ReceivedLineInput{
		{POLineID: lineID, ReceivedQuantity: 10, ReceivedUnitPrice: 6.00},
	}, uuid.New())
	if err != nil {
		t.Fatalf("Receive failed: %v", err)
	}

	if len(varianceRepo.records) != 1 {
		t.Errorf("expected 1 variance record (price mismatch), got %d", len(varianceRepo.records))
	}
}

func TestReceive_BothMismatch_TwoVariances(t *testing.T) {
	poRepo := newMockPORepo()
	lineRepo := newMockPOLineRepo()
	varianceRepo := newMockVarianceRepo()
	landedCostRepo := newMockLandedCostRepo()
	inventoryRepo := newMockInventoryRepo()
	itemRepo := newMockItemRepo()

	svc := newPOService(poRepo, lineRepo, varianceRepo, landedCostRepo, inventoryRepo, itemRepo)

	poID := uuid.New()
	po := &domain.PurchaseOrder{
		ID:        poID,
		Status:    domain.POStatusApproved,
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		Version:   1,
	}
	poRepo.orders[poID] = po

	lineID := uuid.New()
	line := &domain.PurchaseOrderLine{
		ID:               lineID,
		PurchaseOrderID:  poID,
		ItemID:           uuid.New(),
		OrderedQuantity:  10,
		OrderedUnitPrice: 5.00,
	}
	itemRepo.items[line.ItemID] = &domain.Item{ID: line.ItemID, Quantity: 0, Version: 1}
	lineRepo.lines[lineID] = line
	lineRepo.byPO[poID] = append(lineRepo.byPO[poID], lineID)

	// Both qty and price differ.
	err := svc.Receive(context.Background(), poID, []application.ReceivedLineInput{
		{POLineID: lineID, ReceivedQuantity: 8, ReceivedUnitPrice: 6.00},
	}, uuid.New())
	if err != nil {
		t.Fatalf("Receive failed: %v", err)
	}

	if len(varianceRepo.records) != 2 {
		t.Errorf("expected 2 variance records (qty + price), got %d", len(varianceRepo.records))
	}
}

func TestReceive_InventoryAdjusted(t *testing.T) {
	poRepo := newMockPORepo()
	lineRepo := newMockPOLineRepo()
	varianceRepo := newMockVarianceRepo()
	landedCostRepo := newMockLandedCostRepo()
	inventoryRepo := newMockInventoryRepo()
	itemRepo := newMockItemRepo()

	svc := newPOService(poRepo, lineRepo, varianceRepo, landedCostRepo, inventoryRepo, itemRepo)

	poID := uuid.New()
	po := &domain.PurchaseOrder{
		ID:        poID,
		Status:    domain.POStatusApproved,
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		Version:   1,
	}
	poRepo.orders[poID] = po

	lineID := uuid.New()
	line := &domain.PurchaseOrderLine{
		ID:               lineID,
		PurchaseOrderID:  poID,
		ItemID:           uuid.New(),
		OrderedQuantity:  10,
		OrderedUnitPrice: 5.00,
	}
	itemRepo.items[line.ItemID] = &domain.Item{ID: line.ItemID, Quantity: 0, Version: 1}
	lineRepo.lines[lineID] = line
	lineRepo.byPO[poID] = append(lineRepo.byPO[poID], lineID)

	err := svc.Receive(context.Background(), poID, []application.ReceivedLineInput{
		{POLineID: lineID, ReceivedQuantity: 10, ReceivedUnitPrice: 5.00},
	}, uuid.New())
	if err != nil {
		t.Fatalf("Receive failed: %v", err)
	}

	if len(inventoryRepo.adjustments) == 0 {
		t.Error("expected at least 1 inventory adjustment after Receive")
	}
	if inventoryRepo.adjustments[0].QuantityChange != 10 {
		t.Errorf("expected QuantityChange=10, got %d", inventoryRepo.adjustments[0].QuantityChange)
	}
	if len(inventoryRepo.snapshots) != 1 || inventoryRepo.snapshots[0].Quantity != 10 {
		t.Error("expected inventory snapshot with quantity 10")
	}
}

func TestReturn_ReceivedToReturned(t *testing.T) {
	poRepo := newMockPORepo()
	lineRepo := newMockPOLineRepo()
	varianceRepo := newMockVarianceRepo()
	landedCostRepo := newMockLandedCostRepo()
	inventoryRepo := newMockInventoryRepo()
	itemRepo := newMockItemRepo()

	svc := newPOService(poRepo, lineRepo, varianceRepo, landedCostRepo, inventoryRepo, itemRepo)

	poID := uuid.New()
	po := &domain.PurchaseOrder{
		ID:        poID,
		Status:    domain.POStatusReceived,
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		Version:   1,
	}
	poRepo.orders[poID] = po
	lineID := uuid.New()
	itemID := uuid.New()
	qty := 6
	itemRepo.items[itemID] = &domain.Item{ID: itemID, Quantity: qty, Version: 1}
	lineRepo.lines[lineID] = &domain.PurchaseOrderLine{
		ID:              lineID,
		PurchaseOrderID: poID,
		ItemID:          itemID,
		ReceivedQuantity: &qty,
	}
	lineRepo.byPO[poID] = append(lineRepo.byPO[poID], lineID)

	if err := svc.Return(context.Background(), poID); err != nil {
		t.Fatalf("Return failed: %v", err)
	}

	if poRepo.orders[poID].Status != domain.POStatusReturned {
		t.Errorf("expected status=returned, got %v", poRepo.orders[poID].Status)
	}
	if len(inventoryRepo.adjustments) != 1 || inventoryRepo.adjustments[0].QuantityChange != -6 {
		t.Error("expected inventory reversal of -6 on return")
	}
}

func TestVoid_ApprovedToVoided(t *testing.T) {
	poRepo := newMockPORepo()
	lineRepo := newMockPOLineRepo()
	varianceRepo := newMockVarianceRepo()
	landedCostRepo := newMockLandedCostRepo()
	inventoryRepo := newMockInventoryRepo()
	itemRepo := newMockItemRepo()

	svc := newPOService(poRepo, lineRepo, varianceRepo, landedCostRepo, inventoryRepo, itemRepo)

	poID := uuid.New()
	po := &domain.PurchaseOrder{
		ID:        poID,
		Status:    domain.POStatusApproved,
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		Version:   1,
	}
	poRepo.orders[poID] = po

	if err := svc.Void(context.Background(), poID); err != nil {
		t.Fatalf("Void failed: %v", err)
	}

	if poRepo.orders[poID].Status != domain.POStatusVoided {
		t.Errorf("expected status=voided, got %v", poRepo.orders[poID].Status)
	}
}
