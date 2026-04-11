package http

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
	"fitcommerce/internal/http/dto"
	"fitcommerce/internal/security"
)

// ItemHandler handles HTTP requests for the catalog item endpoints.
type ItemHandler struct {
	svc application.ItemService
}

// NewItemHandler creates an ItemHandler backed by the given service.
func NewItemHandler(svc application.ItemService) *ItemHandler {
	return &ItemHandler{svc: svc}
}

// CreateItem handles POST /api/v1/items.
func (h *ItemHandler) CreateItem(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	var req dto.CreateItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", err.Error()))
	}

	item := &domain.Item{
		SKU:               req.SKU,
		Name:              req.Name,
		Description:       req.Description,
		Category:          req.Category,
		Brand:             req.Brand,
		Condition:         domain.ItemCondition(req.Condition),
		UnitPrice:         req.UnitPrice,
		RefundableDeposit: req.RefundableDeposit,
		BillingModel:      domain.BillingModel(req.BillingModel),
		Quantity:          req.Quantity,
		CreatedBy:         user.ID,
	}
	if req.LocationID != nil {
		id, err := uuid.Parse(*req.LocationID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid location_id"))
		}
		item.LocationID = &id
	}

	avail := make([]domain.AvailabilityWindow, len(req.AvailabilityWindows))
	for i, w := range req.AvailabilityWindows {
		avail[i] = domain.AvailabilityWindow{StartTime: w.StartTime, EndTime: w.EndTime}
	}
	blackouts := make([]domain.BlackoutWindow, len(req.BlackoutWindows))
	for i, w := range req.BlackoutWindows {
		blackouts[i] = domain.BlackoutWindow{StartTime: w.StartTime, EndTime: w.EndTime}
	}

	created, err := h.svc.Create(c.Request().Context(), item, avail, blackouts)
	if err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusCreated, dto.SuccessResponse{Data: toItemResponse(created, avail, blackouts)})
}

// ListItems handles GET /api/v1/items.
func (h *ItemHandler) ListItems(c echo.Context) error {
	page, pageSize := parsePagination(c)
	filters := map[string]string{}
	for _, f := range []string{"category", "brand", "condition", "status"} {
		if v := c.QueryParam(f); v != "" {
			filters[f] = v
		}
	}

	items, total, err := h.svc.List(c.Request().Context(), page, pageSize, filters)
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.ItemResponse, len(items))
	for i := range items {
		resp[i] = toItemResponse(&items[i], nil, nil)
	}
	return c.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       resp,
		Pagination: paginationMeta(page, pageSize, total),
	})
}

// GetItem handles GET /api/v1/items/:id.
func (h *ItemHandler) GetItem(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid item id"))
	}

	detail, err := h.svc.GetDetail(c.Request().Context(), id)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{
		Data: toItemResponse(detail.Item, detail.AvailabilityWindows, detail.BlackoutWindows),
	})
}

// UpdateItem handles PUT /api/v1/items/:id.
func (h *ItemHandler) UpdateItem(c echo.Context) error {
	if _, ok := security.GetUserFromContext(c); !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid item id"))
	}

	var req dto.UpdateItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", err.Error()))
	}

	// Fetch existing to merge partial update fields.
	existing, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return HandleDomainError(c, err)
	}

	item := &domain.Item{
		ID:                id,
		SKU:               existing.SKU,
		Name:              existing.Name,
		Description:       existing.Description,
		Category:          existing.Category,
		Brand:             existing.Brand,
		Condition:         existing.Condition,
		UnitPrice:         existing.UnitPrice,
		RefundableDeposit: existing.RefundableDeposit,
		BillingModel:      existing.BillingModel,
		Status:            existing.Status,
		Quantity:          existing.Quantity,
		LocationID:        existing.LocationID,
		CreatedBy:         existing.CreatedBy,
		Version:           req.Version,
	}

	if req.SKU != nil {
		item.SKU = *req.SKU
	}
	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Description != nil {
		item.Description = *req.Description
	}
	if req.Category != nil {
		item.Category = *req.Category
	}
	if req.Brand != nil {
		item.Brand = *req.Brand
	}
	if req.Condition != nil {
		item.Condition = domain.ItemCondition(*req.Condition)
	}
	if req.UnitPrice != nil {
		item.UnitPrice = *req.UnitPrice
	}
	if req.RefundableDeposit != nil {
		item.RefundableDeposit = *req.RefundableDeposit
	}
	if req.BillingModel != nil {
		item.BillingModel = domain.BillingModel(*req.BillingModel)
	}
	if req.Quantity != nil {
		item.Quantity = *req.Quantity
	}
	if req.LocationID != nil {
		lid, err := uuid.Parse(*req.LocationID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid location_id"))
		}
		item.LocationID = &lid
	}

	var avail []domain.AvailabilityWindow
	if req.AvailabilityWindows != nil {
		avail = make([]domain.AvailabilityWindow, len(*req.AvailabilityWindows))
		for i, w := range *req.AvailabilityWindows {
			avail[i] = domain.AvailabilityWindow{StartTime: w.StartTime, EndTime: w.EndTime}
		}
	}
	var blackouts []domain.BlackoutWindow
	if req.BlackoutWindows != nil {
		blackouts = make([]domain.BlackoutWindow, len(*req.BlackoutWindows))
		for i, w := range *req.BlackoutWindows {
			blackouts[i] = domain.BlackoutWindow{StartTime: w.StartTime, EndTime: w.EndTime}
		}
	}

	updated, err := h.svc.Update(c.Request().Context(), item, avail, blackouts)
	if err != nil {
		return HandleDomainError(c, err)
	}

	detail, err := h.svc.GetDetail(c.Request().Context(), updated.ID)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{
		Data: toItemResponse(detail.Item, detail.AvailabilityWindows, detail.BlackoutWindows),
	})
}

// PublishItem handles POST /api/v1/items/:id/publish.
func (h *ItemHandler) PublishItem(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid item id"))
	}
	if err := h.svc.Publish(c.Request().Context(), id); err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: map[string]string{"message": "item published"}})
}

// UnpublishItem handles POST /api/v1/items/:id/unpublish.
func (h *ItemHandler) UnpublishItem(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid item id"))
	}
	if err := h.svc.Unpublish(c.Request().Context(), id); err != nil {
		return HandleDomainError(c, err)
	}
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: map[string]string{"message": "item unpublished"}})
}

// BatchEdit handles POST /api/v1/items/batch-edit.
func (h *ItemHandler) BatchEdit(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	var req dto.BatchEditRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}
	if len(req.Edits) == 0 {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "edits must not be empty"))
	}

	edits := make([]application.BatchEditInput, len(req.Edits))
	for i, e := range req.Edits {
		itemID, err := uuid.Parse(e.ItemID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid item_id in edit row "+strconv.Itoa(i)))
		}

		availability := make([]domain.AvailabilityWindow, len(e.AvailabilityWindows))
		for idx, window := range e.AvailabilityWindows {
			availability[idx] = domain.AvailabilityWindow{
				StartTime: window.StartTime,
				EndTime:   window.EndTime,
			}
		}

		edits[i] = application.BatchEditInput{
			ItemID:              itemID,
			Field:               e.Field,
			NewValue:            e.NewValue,
			AvailabilityWindows: availability,
		}
	}

	job, results, err := h.svc.BatchEdit(c.Request().Context(), user.ID, edits)
	if err != nil {
		return HandleDomainError(c, err)
	}

	respResults := make([]dto.BatchEditResultResponse, len(results))
	for i, r := range results {
		respResults[i] = dto.BatchEditResultResponse{
			ItemID:        r.ItemID.String(),
			Field:         r.Field,
			OldValue:      r.OldValue,
			NewValue:      r.NewValue,
			Success:       r.Success,
			FailureReason: r.FailureReason,
		}
	}
	status := http.StatusOK
	if job.FailureCount > 0 {
		status = http.StatusMultiStatus
	}

	return c.JSON(status, dto.SuccessResponse{Data: dto.BatchEditResponse{
		JobID:        job.ID.String(),
		TotalRows:    job.TotalRows,
		SuccessCount: job.SuccessCount,
		FailureCount: job.FailureCount,
		Results:      respResults,
	}})
}

// --- Private helpers ---

func toItemResponse(item *domain.Item, avail []domain.AvailabilityWindow, blackouts []domain.BlackoutWindow) dto.ItemResponse {
	r := dto.ItemResponse{
		ID:                item.ID.String(),
		SKU:               item.SKU,
		Name:              item.Name,
		Description:       item.Description,
		Category:          item.Category,
		Brand:             item.Brand,
		Condition:         string(item.Condition),
		UnitPrice:         item.UnitPrice,
		RefundableDeposit: item.RefundableDeposit,
		BillingModel:      string(item.BillingModel),
		Status:            string(item.Status),
		Quantity:          item.Quantity,
		CreatedBy:         item.CreatedBy.String(),
		CreatedAt:         item.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         item.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		Version:           item.Version,
	}
	if item.LocationID != nil {
		s := item.LocationID.String()
		r.LocationID = &s
	}
	for _, w := range avail {
		r.AvailabilityWindows = append(r.AvailabilityWindows, dto.AvailabilityWindowResponse{
			ID:        w.ID.String(),
			StartTime: w.StartTime.UTC().Format("2006-01-02T15:04:05Z07:00"),
			EndTime:   w.EndTime.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	for _, w := range blackouts {
		r.BlackoutWindows = append(r.BlackoutWindows, dto.BlackoutWindowResponse{
			ID:        w.ID.String(),
			StartTime: w.StartTime.UTC().Format("2006-01-02T15:04:05Z07:00"),
			EndTime:   w.EndTime.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return r
}

// parsePagination extracts page and page_size from query params with defaults.
func parsePagination(c echo.Context) (int, int) {
	page := 1
	pageSize := 20
	if p, err := strconv.Atoi(c.QueryParam("page")); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(c.QueryParam("page_size")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}
	return page, pageSize
}

// paginationMeta constructs PaginationMeta from page/size/total values.
func paginationMeta(page, pageSize, total int) dto.PaginationMeta {
	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}
	return dto.PaginationMeta{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: total,
		TotalPages: totalPages,
	}
}
