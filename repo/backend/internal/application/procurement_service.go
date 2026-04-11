package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
)

// PurchaseOrderServiceImpl implements PurchaseOrderService.
type PurchaseOrderServiceImpl struct {
	poRepo         store.PurchaseOrderRepository
	lineRepo       store.POLineRepository
	varianceRepo   store.VarianceRepository
	landedCostRepo store.LandedCostRepository
	inventoryRepo  store.InventoryRepository
	itemRepo       store.ItemRepository
	audit          AuditService
	txPool         *pgxpool.Pool
}

// NewPurchaseOrderService creates a PurchaseOrderServiceImpl backed by the given repositories.
func NewPurchaseOrderService(
	poRepo store.PurchaseOrderRepository,
	lineRepo store.POLineRepository,
	varianceRepo store.VarianceRepository,
	landedCostRepo store.LandedCostRepository,
	inventoryRepo store.InventoryRepository,
	itemRepo store.ItemRepository,
	audit AuditService,
	txPool *pgxpool.Pool,
) *PurchaseOrderServiceImpl {
	return &PurchaseOrderServiceImpl{
		poRepo:         poRepo,
		lineRepo:       lineRepo,
		varianceRepo:   varianceRepo,
		landedCostRepo: landedCostRepo,
		inventoryRepo:  inventoryRepo,
		itemRepo:       itemRepo,
		audit:          audit,
		txPool:         txPool,
	}
}

// Create validates lines, persists the purchase order and its lines, and records an audit event.
func (s *PurchaseOrderServiceImpl) Create(ctx context.Context, po *domain.PurchaseOrder, lines []domain.PurchaseOrderLine) (*domain.PurchaseOrder, error) {
	for i := range lines {
		if _, err := s.itemRepo.GetByID(ctx, lines[i].ItemID); err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return nil, fmt.Errorf("procurement_service.Create item %s not found: %w", lines[i].ItemID, domain.ErrNotFound)
			}
			return nil, fmt.Errorf("procurement_service.Create get item: %w", err)
		}
	}

	now := time.Now().UTC()
	po.ID = uuid.New()
	po.Status = domain.POStatusCreated
	po.CreatedAt = now
	po.Version = 1

	var total float64
	for _, line := range lines {
		total += float64(line.OrderedQuantity) * line.OrderedUnitPrice
	}
	po.TotalAmount = total

	if err := withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		if err := s.poRepo.Create(txCtx, po); err != nil {
			return fmt.Errorf("procurement_service.Create po: %w", err)
		}
		for i := range lines {
			lines[i].ID = uuid.New()
			lines[i].PurchaseOrderID = po.ID
			if err := s.lineRepo.Create(txCtx, &lines[i]); err != nil {
				return fmt.Errorf("procurement_service.Create line: %w", err)
			}
		}
		return s.audit.Log(txCtx, "po.created", "purchase_order", po.ID, po.CreatedBy, map[string]interface{}{
			"supplier_id":  po.SupplierID.String(),
			"total_amount": po.TotalAmount,
			"line_count":   len(lines),
		})
	}); err != nil {
		return nil, err
	}
	return po, nil
}

// Get retrieves a purchase order and its lines by ID.
func (s *PurchaseOrderServiceImpl) Get(ctx context.Context, id uuid.UUID) (*domain.PurchaseOrder, []domain.PurchaseOrderLine, error) {
	po, err := s.poRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("procurement_service.Get po: %w", err)
	}

	lines, err := s.lineRepo.ListByPOID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("procurement_service.Get lines: %w", err)
	}

	return po, lines, nil
}

// List returns a paginated list of purchase orders.
func (s *PurchaseOrderServiceImpl) List(ctx context.Context, page, pageSize int) ([]domain.PurchaseOrder, int, error) {
	return s.poRepo.List(ctx, page, pageSize)
}

// Approve transitions a purchase order to the approved status.
func (s *PurchaseOrderServiceImpl) Approve(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) error {
	now := time.Now().UTC()
	return withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		po, err := s.poRepo.GetByID(txCtx, id)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return err
			}
			return fmt.Errorf("procurement_service.Approve get: %w", err)
		}
		if err := domain.TransitionPO(po, domain.POStatusApproved); err != nil {
			return err
		}
		po.ApprovedBy = &approvedBy
		po.ApprovedAt = &now
		if err := s.poRepo.Update(txCtx, po); err != nil {
			return fmt.Errorf("procurement_service.Approve update: %w", err)
		}
		return s.audit.Log(txCtx, "po.approved", "purchase_order", po.ID, approvedBy, map[string]interface{}{
			"approved_at": now.Format(time.RFC3339),
		})
	})
}

// Receive transitions a purchase order to received, records received quantities/prices,
// generates variance records, inventory adjustments, and landed cost entries.
func (s *PurchaseOrderServiceImpl) Receive(ctx context.Context, id uuid.UUID, receivedLines []ReceivedLineInput, actorID uuid.UUID) error {
	now := time.Now().UTC()
	return withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		po, err := s.poRepo.GetByID(txCtx, id)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return err
			}
			return fmt.Errorf("procurement_service.Receive get po: %w", err)
		}
		existingLines, err := s.lineRepo.ListByPOID(txCtx, id)
		if err != nil {
			return fmt.Errorf("procurement_service.Receive list lines: %w", err)
		}
		if err := domain.TransitionPO(po, domain.POStatusReceived); err != nil {
			return err
		}
		po.ReceivedAt = &now
		if err := s.poRepo.Update(txCtx, po); err != nil {
			return fmt.Errorf("procurement_service.Receive update po: %w", err)
		}

		lineMap := make(map[uuid.UUID]*domain.PurchaseOrderLine, len(existingLines))
		for i := range existingLines {
			lineMap[existingLines[i].ID] = &existingLines[i]
		}
		quarter := strconv.Itoa((int(now.Month())-1)/3 + 1)
		period := now.UTC().Format("2006") + "-Q" + quarter
		dueDate := domain.CalculateResolutionDueDate(now)

		for _, receivedLine := range receivedLines {
			line, ok := lineMap[receivedLine.POLineID]
			if !ok {
				return fmt.Errorf("procurement_service.Receive: line %s not found on PO", receivedLine.POLineID)
			}
			line.ReceivedQuantity = &receivedLine.ReceivedQuantity
			line.ReceivedUnitPrice = &receivedLine.ReceivedUnitPrice
			if err := s.lineRepo.Update(txCtx, line); err != nil {
				return fmt.Errorf("procurement_service.Receive update line: %w", err)
			}

			if receivedLine.ReceivedQuantity != line.OrderedQuantity {
				var varType domain.VarianceType
				if receivedLine.ReceivedQuantity < line.OrderedQuantity {
					varType = domain.VarianceTypeShortage
				} else {
					varType = domain.VarianceTypeOverage
				}
				expected := float64(line.OrderedQuantity)
				actual := float64(receivedLine.ReceivedQuantity)
				diff := actual - expected
				if err := s.varianceRepo.Create(txCtx, &domain.VarianceRecord{
					ID:                uuid.New(),
					POLineID:          line.ID,
					Type:              varType,
					ExpectedValue:     expected,
					ActualValue:       actual,
					DifferenceAmount:  diff,
					Status:            domain.VarianceStatusOpen,
					ResolutionDueDate: dueDate,
					ResolvedAt:        nil,
					ResolutionNotes:   "",
					CreatedAt:         now,
				}); err != nil {
					return fmt.Errorf("procurement_service.Receive create quantity variance: %w", err)
				}
			}

			if math.Abs(receivedLine.ReceivedUnitPrice-line.OrderedUnitPrice) > 0.0001 {
				expected := line.OrderedUnitPrice
				actual := receivedLine.ReceivedUnitPrice
				diff := actual - expected
				if err := s.varianceRepo.Create(txCtx, &domain.VarianceRecord{
					ID:                uuid.New(),
					POLineID:          line.ID,
					Type:              domain.VarianceTypePriceDifference,
					ExpectedValue:     expected,
					ActualValue:       actual,
					DifferenceAmount:  diff,
					Status:            domain.VarianceStatusOpen,
					ResolutionDueDate: dueDate,
					ResolvedAt:        nil,
					ResolutionNotes:   "",
					CreatedAt:         now,
				}); err != nil {
					return fmt.Errorf("procurement_service.Receive create price variance: %w", err)
				}
			}

			if err := s.inventoryCoordinator().applyChange(txCtx, line.ItemID, receivedLine.ReceivedQuantity, "po-receipt", actorID, now); err != nil {
				return fmt.Errorf("procurement_service.Receive inventory: %w", err)
			}
			rawAmount := float64(receivedLine.ReceivedQuantity) * receivedLine.ReceivedUnitPrice
			if err := s.landedCostRepo.Create(txCtx, &domain.LandedCostEntry{
				ID:               uuid.New(),
				ItemID:           line.ItemID,
				PurchaseOrderID:  po.ID,
				POLineID:         line.ID,
				Period:           period,
				CostComponent:    "unit_cost",
				RawAmount:        rawAmount,
				AllocatedAmount:  rawAmount,
				AllocationMethod: "direct",
				CreatedAt:        now,
			}); err != nil {
				return fmt.Errorf("procurement_service.Receive create landed cost: %w", err)
			}
		}

		if err := s.audit.Log(txCtx, "po.received", "purchase_order", po.ID, actorID, map[string]interface{}{
			"received_at": now.Format(time.RFC3339),
			"line_count":  len(receivedLines),
		}); err != nil {
			slog.Default().Warn("audit log failed", "event", "po.received", "error", err)
			return err
		}
		return nil
	})
}

// Return transitions a purchase order to the returned status.
func (s *PurchaseOrderServiceImpl) Return(ctx context.Context, id uuid.UUID, performedBy uuid.UUID) error {
	return withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		po, err := s.poRepo.GetByID(txCtx, id)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return err
			}
			return fmt.Errorf("procurement_service.Return get: %w", err)
		}
		if err := domain.TransitionPO(po, domain.POStatusReturned); err != nil {
			return err
		}
		lines, err := s.lineRepo.ListByPOID(txCtx, id)
		if err != nil {
			return fmt.Errorf("procurement_service.Return lines: %w", err)
		}
		now := time.Now().UTC()
		for _, line := range lines {
			if line.ReceivedQuantity != nil && *line.ReceivedQuantity > 0 {
				if err := s.inventoryCoordinator().applyChange(txCtx, line.ItemID, -*line.ReceivedQuantity, "po-return", performedBy, now); err != nil {
					return err
				}
			}
		}
		if err := s.poRepo.Update(txCtx, po); err != nil {
			return fmt.Errorf("procurement_service.Return update: %w", err)
		}
		if err := s.audit.Log(txCtx, "po.returned", "purchase_order", po.ID, performedBy, nil); err != nil {
			slog.Default().Warn("audit log failed", "event", "po.returned", "error", err)
			return err
		}
		return nil
	})
}

// Void transitions a purchase order to the voided status.
func (s *PurchaseOrderServiceImpl) Void(ctx context.Context, id uuid.UUID, performedBy uuid.UUID) error {
	return withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		po, err := s.poRepo.GetByID(txCtx, id)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return err
			}
			return fmt.Errorf("procurement_service.Void get: %w", err)
		}
		if err := domain.TransitionPO(po, domain.POStatusVoided); err != nil {
			return err
		}
		if err := s.poRepo.Update(txCtx, po); err != nil {
			return fmt.Errorf("procurement_service.Void update: %w", err)
		}
		if err := s.audit.Log(txCtx, "po.voided", "purchase_order", po.ID, performedBy, nil); err != nil {
			slog.Default().Warn("audit log failed", "event", "po.voided", "error", err)
			return err
		}
		return nil
	})
}

func (s *PurchaseOrderServiceImpl) inventoryCoordinator() inventoryCoordinator {
	return inventoryCoordinator{
		items:     s.itemRepo,
		inventory: s.inventoryRepo,
	}
}

// Compile-time interface assertion.
var _ PurchaseOrderService = (*PurchaseOrderServiceImpl)(nil)
