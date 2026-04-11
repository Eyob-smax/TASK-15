package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
	"fitcommerce/internal/store/postgres"
)

func withOptionalTransaction(ctx context.Context, pool *pgxpool.Pool, fn func(context.Context) error) error {
	if pool == nil {
		return fn(ctx)
	}
	return postgres.WithTransaction(ctx, pool, fn)
}

type inventoryCoordinator struct {
	items        store.ItemRepository
	inventory    store.InventoryRepository
	availability store.AvailabilityWindowRepository
	blackouts    store.BlackoutWindowRepository
}

func (c inventoryCoordinator) ensureReservable(ctx context.Context, itemID uuid.UUID, quantity int, now time.Time) (*domain.Item, error) {
	item, err := c.items.GetByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item.Status != domain.ItemStatusPublished {
		return nil, &domain.ErrValidation{Field: "item_id", Message: "item is not available for ordering"}
	}
	if quantity < 1 {
		return nil, &domain.ErrValidation{Field: "quantity", Message: "quantity must be at least 1"}
	}
	if item.Quantity < quantity {
		return nil, &domain.ErrValidation{Field: "quantity", Message: "requested quantity exceeds available stock"}
	}

	if c.availability != nil {
		windows, err := c.availability.ListByItemID(ctx, itemID)
		if err != nil {
			return nil, err
		}
		if len(windows) > 0 && !timeInAvailabilityWindows(now, windows) {
			return nil, &domain.ErrValidation{Field: "item_id", Message: "item is outside its availability window"}
		}
	}

	if c.blackouts != nil {
		blackouts, err := c.blackouts.ListByItemID(ctx, itemID)
		if err != nil {
			return nil, err
		}
		if timeInBlackoutWindows(now, blackouts) {
			return nil, &domain.ErrValidation{Field: "item_id", Message: "item is unavailable during a blackout window"}
		}
	}

	return item, nil
}

func (c inventoryCoordinator) applyChange(ctx context.Context, itemID uuid.UUID, quantityChange int, reason string, actorID uuid.UUID, now time.Time) error {
	item, err := c.items.GetByID(ctx, itemID)
	if err != nil {
		return err
	}

	newQty := item.Quantity + quantityChange
	if newQty < 0 {
		return &domain.ErrValidation{Field: "quantity", Message: "inventory change would make quantity negative"}
	}

	item.Quantity = newQty
	item.UpdatedAt = now
	item.Version++
	if err := c.items.Update(ctx, item); err != nil {
		return err
	}

	if err := c.inventory.CreateAdjustment(ctx, &domain.InventoryAdjustment{
		ID:             uuid.New(),
		ItemID:         itemID,
		QuantityChange: quantityChange,
		Reason:         reason,
		CreatedBy:      actorID,
		CreatedAt:      now,
	}); err != nil {
		return err
	}

	return c.inventory.CreateSnapshot(ctx, &domain.InventorySnapshot{
		ID:         uuid.New(),
		ItemID:     itemID,
		Quantity:   newQty,
		LocationID: item.LocationID,
		RecordedAt: now,
	})
}

func timeInAvailabilityWindows(now time.Time, windows []domain.AvailabilityWindow) bool {
	for _, window := range windows {
		if !now.Before(window.StartTime) && now.Before(window.EndTime) {
			return true
		}
	}
	return false
}

func timeInBlackoutWindows(now time.Time, windows []domain.BlackoutWindow) bool {
	for _, window := range windows {
		if !now.Before(window.StartTime) && now.Before(window.EndTime) {
			return true
		}
	}
	return false
}
