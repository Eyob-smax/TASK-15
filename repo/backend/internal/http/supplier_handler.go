package http

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
	"fitcommerce/internal/http/dto"
)

// SupplierHandler handles HTTP requests for supplier endpoints.
type SupplierHandler struct {
	svc application.SupplierService
}

// NewSupplierHandler creates a SupplierHandler backed by the given service.
func NewSupplierHandler(svc application.SupplierService) *SupplierHandler {
	return &SupplierHandler{svc: svc}
}

// CreateSupplier handles POST /api/v1/suppliers.
func (h *SupplierHandler) CreateSupplier(c echo.Context) error {
	var req dto.CreateSupplierRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", err.Error()))
	}

	supplier := domain.Supplier{
		Name:         req.Name,
		ContactName:  req.ContactName,
		ContactEmail: req.ContactEmail,
		ContactPhone: req.ContactPhone,
		Address:      req.Address,
		IsActive:     true,
	}

	created, err := h.svc.Create(c.Request().Context(), &supplier)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusCreated, dto.SuccessResponse{Data: toSupplierResponse(created)})
}

// ListSuppliers handles GET /api/v1/suppliers.
func (h *SupplierHandler) ListSuppliers(c echo.Context) error {
	page, pageSize := parsePagination(c)

	suppliers, total, err := h.svc.List(c.Request().Context(), page, pageSize)
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.SupplierResponse, len(suppliers))
	for i := range suppliers {
		resp[i] = toSupplierResponse(&suppliers[i])
	}

	return c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       resp,
		Pagination: paginationMeta(page, pageSize, total),
	})
}

// GetSupplier handles GET /api/v1/suppliers/:id.
func (h *SupplierHandler) GetSupplier(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid supplier id"))
	}

	supplier, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: toSupplierResponse(supplier)})
}

// UpdateSupplier handles PUT /api/v1/suppliers/:id.
func (h *SupplierHandler) UpdateSupplier(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid supplier id"))
	}

	existing, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return HandleDomainError(c, err)
	}

	var req dto.CreateSupplierRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", err.Error()))
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.ContactName != "" {
		existing.ContactName = req.ContactName
	}
	if req.ContactEmail != "" {
		existing.ContactEmail = req.ContactEmail
	}
	if req.ContactPhone != "" {
		existing.ContactPhone = req.ContactPhone
	}
	if req.Address != "" {
		existing.Address = req.Address
	}

	if err := h.svc.Update(c.Request().Context(), existing); err != nil {
		return HandleDomainError(c, err)
	}

	updated, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: toSupplierResponse(updated)})
}

// --- Private helpers ---

func toSupplierResponse(s *domain.Supplier) dto.SupplierResponse {
	return dto.SupplierResponse{
		ID:           s.ID.String(),
		Name:         s.Name,
		ContactName:  s.ContactName,
		ContactEmail: s.ContactEmail,
		ContactPhone: s.ContactPhone,
		Address:      s.Address,
		IsActive:     s.IsActive,
		CreatedAt:    s.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    s.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
