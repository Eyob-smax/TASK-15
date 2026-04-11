package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
)

// LandedCostServiceImpl implements LandedCostService.
type LandedCostServiceImpl struct {
	repo store.LandedCostRepository
}

// NewLandedCostService creates a LandedCostServiceImpl backed by the given repository.
func NewLandedCostService(repo store.LandedCostRepository) *LandedCostServiceImpl {
	return &LandedCostServiceImpl{repo: repo}
}

// GetSummary returns landed cost entries for a specific item and period.
func (s *LandedCostServiceImpl) GetSummary(ctx context.Context, itemID uuid.UUID, period string) ([]domain.LandedCostEntry, error) {
	entries, err := s.repo.ListByItemAndPeriod(ctx, itemID, period)
	if err != nil {
		return nil, fmt.Errorf("landed_cost_service.GetSummary: %w", err)
	}
	return entries, nil
}

// GetByPOID returns landed cost entries for a specific purchase order.
func (s *LandedCostServiceImpl) GetByPOID(ctx context.Context, poID uuid.UUID) ([]domain.LandedCostEntry, error) {
	entries, err := s.repo.ListByPOID(ctx, poID)
	if err != nil {
		return nil, fmt.Errorf("landed_cost_service.GetByPOID: %w", err)
	}
	return entries, nil
}

// Compile-time interface assertion.
var _ LandedCostService = (*LandedCostServiceImpl)(nil)
