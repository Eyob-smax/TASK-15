package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
)

// VarianceServiceImpl implements VarianceService.
type VarianceServiceImpl struct {
	varianceRepo store.VarianceRepository
	lineRepo     store.POLineRepository
	itemRepo     store.ItemRepository
	inventory    store.InventoryRepository
	audit        AuditService
	txPool       *pgxpool.Pool
}

// NewVarianceService creates a VarianceServiceImpl backed by the given repository and audit service.
func NewVarianceService(
	varianceRepo store.VarianceRepository,
	lineRepo store.POLineRepository,
	itemRepo store.ItemRepository,
	inventory store.InventoryRepository,
	auditSvc AuditService,
	txPool *pgxpool.Pool,
) *VarianceServiceImpl {
	return &VarianceServiceImpl{
		varianceRepo: varianceRepo,
		lineRepo:     lineRepo,
		itemRepo:     itemRepo,
		inventory:    inventory,
		audit:        auditSvc,
		txPool:       txPool,
	}
}

// Get retrieves a variance record by ID.
func (s *VarianceServiceImpl) Get(ctx context.Context, id uuid.UUID) (*domain.VarianceRecord, error) {
	record, err := s.varianceRepo.GetByID(ctx, id)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, err
		}
		return nil, fmt.Errorf("variance_service.Get: %w", err)
	}
	return record, nil
}

// List returns a paginated list of variance records, optionally filtered by status.
func (s *VarianceServiceImpl) List(ctx context.Context, status *domain.VarianceStatus, page, pageSize int) ([]domain.VarianceRecord, int, error) {
	return s.varianceRepo.List(ctx, status, page, pageSize)
}

// Resolve marks a variance record as resolved using either an inventory
// adjustment or a stock return derived from the linked PO line.
func (s *VarianceServiceImpl) Resolve(ctx context.Context, id uuid.UUID, action, resolutionNotes string, quantityChange *int, performedBy uuid.UUID) error {
	return withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		record, err := s.varianceRepo.GetByID(txCtx, id)
		if err != nil {
			if err == domain.ErrNotFound {
				return err
			}
			return fmt.Errorf("variance_service.Resolve get: %w", err)
		}

		if record.Status != domain.VarianceStatusOpen && record.Status != domain.VarianceStatusEscalated {
			return &domain.ErrInvalidTransition{
				Entity: "variance",
				From:   string(record.Status),
				To:     string(domain.VarianceStatusResolved),
			}
		}

		line, err := s.lineRepo.GetByID(txCtx, record.POLineID)
		if err != nil {
			return fmt.Errorf("variance_service.Resolve line: %w", err)
		}

		switch action {
		case "adjustment":
			if quantityChange == nil || *quantityChange == 0 {
				return &domain.ErrValidation{Field: "quantity_change", Message: "quantity_change is required for adjustment resolutions"}
			}
			if err := s.inventoryCoordinator().applyChange(txCtx, line.ItemID, *quantityChange, "variance adjustment", performedBy, time.Now().UTC()); err != nil {
				return err
			}
		case "return":
			if line.ReceivedQuantity == nil || *line.ReceivedQuantity < 1 {
				return &domain.ErrValidation{Field: "action", Message: "return resolution requires a received quantity to reverse"}
			}
			reversal := -*line.ReceivedQuantity
			quantityChange = &reversal
			if err := s.inventoryCoordinator().applyChange(txCtx, line.ItemID, reversal, "variance return", performedBy, time.Now().UTC()); err != nil {
				return err
			}
		default:
			return &domain.ErrValidation{Field: "action", Message: "unsupported resolution action"}
		}

		now := time.Now().UTC()
		record.Status = domain.VarianceStatusResolved
		record.ResolutionAction = action
		record.ResolutionNotes = resolutionNotes
		record.QuantityChange = quantityChange
		record.ResolvedAt = &now
		if err := s.varianceRepo.Update(txCtx, record); err != nil {
			return fmt.Errorf("variance_service.Resolve update: %w", err)
		}

		if err := s.audit.Log(txCtx, "variance.resolved", "variance", record.ID, performedBy, map[string]interface{}{
			"po_line_id":         record.POLineID,
			"resolution_action":  action,
			"resolution_notes":   resolutionNotes,
			"quantity_change":    quantityChange,
		}); err != nil {
			slog.Default().Warn("audit log failed", "event", "variance.resolved", "error", err)
			return err
		}
		return nil
	})
}

// EscalateOverdue transitions all open, overdue variance records to the
// escalated status. Returns the number of records successfully escalated.
func (s *VarianceServiceImpl) EscalateOverdue(ctx context.Context) (int, error) {
	openStatus := domain.VarianceStatusOpen
	records, _, err := s.varianceRepo.List(ctx, &openStatus, 1, 1000)
	if err != nil {
		return 0, fmt.Errorf("variance_service.EscalateOverdue list: %w", err)
	}

	now := time.Now().UTC()
	count := 0
	for i := range records {
		if !records[i].IsOverdue(now) {
			continue
		}
		records[i].Status = domain.VarianceStatusEscalated
		if err := s.varianceRepo.Update(ctx, &records[i]); err != nil {
			slog.Default().Warn("variance escalation update failed", "id", records[i].ID, "error", err)
			continue
		}
		_ = s.audit.Log(ctx, "variance.escalated", "variance", records[i].ID, domain.SystemActorID, map[string]interface{}{
			"po_line_id": records[i].POLineID,
		})
		count++
	}
	return count, nil
}

// Compile-time interface assertion.
var _ VarianceService = (*VarianceServiceImpl)(nil)

func (s *VarianceServiceImpl) inventoryCoordinator() inventoryCoordinator {
	return inventoryCoordinator{
		items:     s.itemRepo,
		inventory: s.inventory,
	}
}
