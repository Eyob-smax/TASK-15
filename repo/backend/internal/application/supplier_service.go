package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
)

// SupplierServiceImpl implements SupplierService.
type SupplierServiceImpl struct {
	suppliers store.SupplierRepository
	audit     AuditService
}

// NewSupplierService creates a SupplierServiceImpl backed by the given repository.
func NewSupplierService(
	suppliers store.SupplierRepository,
	audit AuditService,
) *SupplierServiceImpl {
	return &SupplierServiceImpl{
		suppliers: suppliers,
		audit:     audit,
	}
}

// Create validates and persists a new supplier record.
func (s *SupplierServiceImpl) Create(ctx context.Context, supplier *domain.Supplier) (*domain.Supplier, error) {
	if supplier.Name == "" {
		return nil, &domain.ErrValidation{Field: "name", Message: "name is required"}
	}

	now := time.Now().UTC()
	supplier.ID = uuid.New()
	supplier.CreatedAt = now
	supplier.UpdatedAt = now

	if err := s.suppliers.Create(ctx, supplier); err != nil {
		return nil, fmt.Errorf("supplier_service.Create: %w", err)
	}

	if err := s.audit.Log(ctx, "supplier.created", "supplier", supplier.ID, domain.SystemActorID, map[string]interface{}{
		"name": supplier.Name,
	}); err != nil {
		slog.Default().Warn("audit log failed", "event", "supplier.created", "error", err)
	}

	return supplier, nil
}

// Get retrieves a supplier by ID.
func (s *SupplierServiceImpl) Get(ctx context.Context, id uuid.UUID) (*domain.Supplier, error) {
	supplier, err := s.suppliers.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("supplier_service.Get: %w", err)
	}
	return supplier, nil
}

// List returns a paginated list of suppliers.
func (s *SupplierServiceImpl) List(ctx context.Context, page, pageSize int) ([]domain.Supplier, int, error) {
	return s.suppliers.List(ctx, page, pageSize)
}

// Update applies changes to an existing supplier record.
func (s *SupplierServiceImpl) Update(ctx context.Context, supplier *domain.Supplier) error {
	existing, err := s.suppliers.GetByID(ctx, supplier.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return err
		}
		return fmt.Errorf("supplier_service.Update get: %w", err)
	}

	existing.Name = supplier.Name
	existing.ContactName = supplier.ContactName
	existing.ContactEmail = supplier.ContactEmail
	existing.ContactPhone = supplier.ContactPhone
	existing.Address = supplier.Address
	existing.IsActive = supplier.IsActive
	existing.UpdatedAt = time.Now().UTC()

	if err := s.suppliers.Update(ctx, existing); err != nil {
		return fmt.Errorf("supplier_service.Update: %w", err)
	}

	if err := s.audit.Log(ctx, "supplier.updated", "supplier", existing.ID, domain.SystemActorID, map[string]interface{}{
		"name": existing.Name,
	}); err != nil {
		slog.Default().Warn("audit log failed", "event", "supplier.updated", "error", err)
	}

	return nil
}

// Compile-time interface assertion.
var _ SupplierService = (*SupplierServiceImpl)(nil)
