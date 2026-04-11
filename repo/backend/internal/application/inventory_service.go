package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
)

// InventoryServiceImpl implements InventoryService.
type InventoryServiceImpl struct {
	inv   store.InventoryRepository
	bins  store.WarehouseBinRepository
	items store.ItemRepository
	audit AuditService
	txPool *pgxpool.Pool
}

// NewInventoryService creates an InventoryServiceImpl backed by the given repositories.
func NewInventoryService(
	inv store.InventoryRepository,
	bins store.WarehouseBinRepository,
	items store.ItemRepository,
	audit AuditService,
	txPool *pgxpool.Pool,
) *InventoryServiceImpl {
	return &InventoryServiceImpl{inv: inv, bins: bins, items: items, audit: audit, txPool: txPool}
}

func (s *InventoryServiceImpl) GetSnapshots(ctx context.Context, itemID *uuid.UUID, locationID *uuid.UUID) ([]domain.InventorySnapshot, error) {
	return s.inv.ListSnapshots(ctx, itemID, locationID)
}

func (s *InventoryServiceImpl) CreateAdjustment(ctx context.Context, adjustment *domain.InventoryAdjustment) (*domain.InventoryAdjustment, error) {
	if adjustment.ID == uuid.Nil {
		adjustment.ID = uuid.New()
	}
	if adjustment.CreatedAt.IsZero() {
		adjustment.CreatedAt = time.Now().UTC()
	}

	if err := withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		item, err := s.items.GetByID(txCtx, adjustment.ItemID)
		if err != nil {
			return err
		}

		newQty := item.Quantity + adjustment.QuantityChange
		if newQty < 0 {
			return &domain.ErrValidation{Field: "quantity_change", Message: "inventory change would make quantity negative"}
		}

		item.Quantity = newQty
		item.UpdatedAt = adjustment.CreatedAt
		item.Version++
		if err := s.items.Update(txCtx, item); err != nil {
			return err
		}

		if err := s.inv.CreateAdjustment(txCtx, adjustment); err != nil {
			return err
		}

		return s.inv.CreateSnapshot(txCtx, &domain.InventorySnapshot{
			ID:         uuid.New(),
			ItemID:     adjustment.ItemID,
			Quantity:   newQty,
			LocationID: item.LocationID,
			RecordedAt: adjustment.CreatedAt,
		})
	}); err != nil {
		return nil, err
	}

	_ = s.audit.Log(ctx, "inventory.adjusted", "item", adjustment.ItemID, adjustment.CreatedBy, map[string]interface{}{
		"quantity_change": adjustment.QuantityChange,
		"reason":         adjustment.Reason,
	})
	return adjustment, nil
}

func (s *InventoryServiceImpl) ListAdjustments(ctx context.Context, itemID *uuid.UUID, page, pageSize int) ([]domain.InventoryAdjustment, int, error) {
	return s.inv.ListAdjustments(ctx, itemID, page, pageSize)
}

func (s *InventoryServiceImpl) CreateWarehouseBin(ctx context.Context, bin *domain.WarehouseBin) (*domain.WarehouseBin, error) {
	now := time.Now().UTC()
	bin.ID = uuid.New()
	bin.CreatedAt = now
	bin.UpdatedAt = now
	if err := s.bins.Create(ctx, bin); err != nil {
		return nil, err
	}
	return bin, nil
}

func (s *InventoryServiceImpl) GetWarehouseBin(ctx context.Context, id uuid.UUID) (*domain.WarehouseBin, error) {
	return s.bins.GetByID(ctx, id)
}

func (s *InventoryServiceImpl) ListWarehouseBins(ctx context.Context, locationID *uuid.UUID, page, pageSize int) ([]domain.WarehouseBin, int, error) {
	return s.bins.List(ctx, locationID, page, pageSize)
}
