package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
)

// LocationServiceImpl implements LocationService.
type LocationServiceImpl struct {
	repo store.LocationRepository
}

// NewLocationService creates a LocationServiceImpl backed by the given repository.
func NewLocationService(repo store.LocationRepository) *LocationServiceImpl {
	return &LocationServiceImpl{repo: repo}
}

// Create validates and persists a new location.
func (s *LocationServiceImpl) Create(ctx context.Context, loc *domain.Location) (*domain.Location, error) {
	if loc.Name == "" {
		return nil, &domain.ErrValidation{Field: "name", Message: "name is required"}
	}
	now := time.Now().UTC()
	loc.ID = uuid.New()
	loc.IsActive = true
	loc.CreatedAt = now
	loc.UpdatedAt = now
	if err := s.repo.Create(ctx, loc); err != nil {
		return nil, fmt.Errorf("location_service.Create: %w", err)
	}
	return loc, nil
}

// GetByID retrieves a location by ID.
func (s *LocationServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.Location, error) {
	return s.repo.GetByID(ctx, id)
}

// List returns a paginated list of locations.
func (s *LocationServiceImpl) List(ctx context.Context, page, pageSize int) ([]domain.Location, int, error) {
	return s.repo.List(ctx, page, pageSize)
}

// Compile-time interface assertion.
var _ LocationService = (*LocationServiceImpl)(nil)
