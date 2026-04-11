package application

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/store"
)

// ItemServiceImpl implements ItemService.
type ItemServiceImpl struct {
	items    store.ItemRepository
	avail    store.AvailabilityWindowRepository
	blackout store.BlackoutWindowRepository
	batches  store.BatchEditRepository
	audit    AuditService
	txPool   *pgxpool.Pool
}

// NewItemService creates an ItemServiceImpl backed by the given repositories.
func NewItemService(
	items store.ItemRepository,
	avail store.AvailabilityWindowRepository,
	blackout store.BlackoutWindowRepository,
	batches store.BatchEditRepository,
	audit AuditService,
	txPool *pgxpool.Pool,
) *ItemServiceImpl {
	return &ItemServiceImpl{
		items:    items,
		avail:    avail,
		blackout: blackout,
		batches:  batches,
		audit:    audit,
		txPool:   txPool,
	}
}

func (s *ItemServiceImpl) Create(ctx context.Context, item *domain.Item, availability []domain.AvailabilityWindow, blackouts []domain.BlackoutWindow) (*domain.Item, error) {
	item.ApplyDepositDefault()
	if errs := item.Validate(); len(errs) > 0 {
		return nil, &domain.ErrValidation{Field: errs[0].Field, Message: errs[0].Message}
	}
	if err := validateAvailabilityWindows(availability); err != nil {
		return nil, err
	}
	if err := validateBlackoutWindows(blackouts); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	item.ID = uuid.New()
	item.Status = domain.ItemStatusDraft
	item.Version = 1
	item.CreatedAt = now
	item.UpdatedAt = now

	if err := withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		if err := s.items.Create(txCtx, item); err != nil {
			return err
		}

		for i := range availability {
			availability[i].ID = uuid.New()
			availability[i].ItemID = item.ID
			if err := s.avail.Create(txCtx, &availability[i]); err != nil {
				return err
			}
		}
		for i := range blackouts {
			blackouts[i].ID = uuid.New()
			blackouts[i].ItemID = item.ID
			if err := s.blackout.Create(txCtx, &blackouts[i]); err != nil {
				return err
			}
		}

		return s.audit.Log(txCtx, "item.created", "item", item.ID, item.CreatedBy, map[string]interface{}{
			"name":   item.Name,
			"status": string(item.Status),
		})
	}); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ItemServiceImpl) Get(ctx context.Context, id uuid.UUID) (*domain.Item, error) {
	return s.items.GetByID(ctx, id)
}

func (s *ItemServiceImpl) GetDetail(ctx context.Context, id uuid.UUID) (*ItemDetail, error) {
	item, err := s.items.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	availability, err := s.avail.ListByItemID(ctx, id)
	if err != nil {
		return nil, err
	}

	blackouts, err := s.blackout.ListByItemID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &ItemDetail{
		Item:                item,
		AvailabilityWindows: availability,
		BlackoutWindows:     blackouts,
	}, nil
}

func (s *ItemServiceImpl) List(ctx context.Context, page, pageSize int, filters map[string]string) ([]domain.Item, int, error) {
	return s.items.List(ctx, filters, page, pageSize)
}

func (s *ItemServiceImpl) Update(ctx context.Context, item *domain.Item, availability []domain.AvailabilityWindow, blackouts []domain.BlackoutWindow) (*domain.Item, error) {
	existing, err := s.items.GetByID(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	if existing.Version != item.Version {
		return nil, &domain.ErrConflict{Entity: "item", Message: "version mismatch"}
	}

	item.ApplyDepositDefault()
	if errs := item.Validate(); len(errs) > 0 {
		return nil, &domain.ErrValidation{Field: errs[0].Field, Message: errs[0].Message}
	}
	if availability != nil {
		if err := validateAvailabilityWindows(availability); err != nil {
			return nil, err
		}
	}
	if blackouts != nil {
		if err := validateBlackoutWindows(blackouts); err != nil {
			return nil, err
		}
	}

	item.Version++
	item.UpdatedAt = time.Now().UTC()
	item.CreatedAt = existing.CreatedAt
	item.CreatedBy = existing.CreatedBy
	item.Status = existing.Status

	if err := withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
		if err := s.items.Update(txCtx, item); err != nil {
			return err
		}

		// nil means "preserve existing windows"; non-nil (even empty) means "replace".
		if availability != nil {
			if err := s.avail.DeleteByItemID(txCtx, item.ID); err != nil {
				return err
			}
			for i := range availability {
				availability[i].ID = uuid.New()
				availability[i].ItemID = item.ID
				if err := s.avail.Create(txCtx, &availability[i]); err != nil {
					return err
				}
			}
		}

		if blackouts != nil {
			if err := s.blackout.DeleteByItemID(txCtx, item.ID); err != nil {
				return err
			}
			for i := range blackouts {
				blackouts[i].ID = uuid.New()
				blackouts[i].ItemID = item.ID
				if err := s.blackout.Create(txCtx, &blackouts[i]); err != nil {
					return err
				}
			}
		}

		return s.audit.Log(txCtx, "item.updated", "item", item.ID, item.CreatedBy, map[string]interface{}{
			"version": item.Version,
		})
	}); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ItemServiceImpl) Publish(ctx context.Context, id uuid.UUID) error {
	item, err := s.items.GetByID(ctx, id)
	if err != nil {
		return err
	}
	availWindows, err := s.avail.ListByItemID(ctx, id)
	if err != nil {
		return err
	}
	blackoutWindows, err := s.blackout.ListByItemID(ctx, id)
	if err != nil {
		return err
	}

	if blocked := domain.ValidateItemForPublish(*item, availWindows, blackoutWindows); blocked != nil {
		return blocked
	}
	if item.Status == domain.ItemStatusPublished {
		return &domain.ErrConflict{Entity: "item", Message: "item is already published"}
	}

	item.Status = domain.ItemStatusPublished
	item.Version++
	item.UpdatedAt = time.Now().UTC()

	if err := s.items.Update(ctx, item); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, "item.published", "item", item.ID, item.CreatedBy, nil)
	return nil
}

func (s *ItemServiceImpl) Unpublish(ctx context.Context, id uuid.UUID) error {
	item, err := s.items.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if item.Status != domain.ItemStatusPublished {
		return &domain.ErrInvalidTransition{Entity: "item", From: string(item.Status), To: "draft"}
	}

	item.Status = domain.ItemStatusDraft
	item.Version++
	item.UpdatedAt = time.Now().UTC()

	if err := s.items.Update(ctx, item); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, "item.unpublished", "item", item.ID, item.CreatedBy, nil)
	return nil
}

func (s *ItemServiceImpl) BatchEdit(ctx context.Context, createdBy uuid.UUID, edits []BatchEditInput) (*domain.BatchEditJob, []domain.BatchEditResult, error) {
	now := time.Now().UTC()
	job := &domain.BatchEditJob{
		ID:        uuid.New(),
		CreatedBy: createdBy,
		CreatedAt: now,
		TotalRows: len(edits),
	}
	if err := s.batches.CreateJob(ctx, job); err != nil {
		return nil, nil, err
	}

	results := make([]domain.BatchEditResult, 0, len(edits))
	successCount := 0
	failureCount := 0

	for _, edit := range edits {
		result := domain.BatchEditResult{
			ID:      uuid.New(),
			BatchID: job.ID,
			ItemID:  edit.ItemID,
			Field:   edit.Field,
		}

		rowErr := withOptionalTransaction(ctx, s.txPool, func(txCtx context.Context) error {
			item, err := s.items.GetByID(txCtx, edit.ItemID)
			if err != nil {
				return err
			}

			currentAvailability, err := s.avail.ListByItemID(txCtx, item.ID)
			if err != nil {
				return err
			}
			currentBlackouts, err := s.blackout.ListByItemID(txCtx, item.ID)
			if err != nil {
				return err
			}

			nextAvailability := cloneAvailabilityWindows(currentAvailability)
			switch edit.Field {
			case "availability_windows":
				result.OldValue = serializeAvailabilityWindows(currentAvailability)
				result.NewValue = serializeAvailabilityWindows(edit.AvailabilityWindows)
				nextAvailability = cloneAvailabilityWindows(edit.AvailabilityWindows)
			default:
				result.OldValue = currentFieldValue(item, edit.Field)
				if edit.NewValue == nil {
					return &domain.ErrValidation{Field: edit.Field, Message: "new_value is required"}
				}
				result.NewValue = *edit.NewValue
				if err := applyFieldChange(item, edit.Field, *edit.NewValue); err != nil {
					return err
				}
			}

			if err := validateBatchEditState(item, nextAvailability, currentBlackouts); err != nil {
				return err
			}

			item.Version++
			item.UpdatedAt = now
			if err := s.items.Update(txCtx, item); err != nil {
				return err
			}

			if edit.Field == "availability_windows" {
				if err := s.avail.DeleteByItemID(txCtx, item.ID); err != nil {
					return err
				}
				for i := range nextAvailability {
					nextAvailability[i].ID = uuid.New()
					nextAvailability[i].ItemID = item.ID
					if err := s.avail.Create(txCtx, &nextAvailability[i]); err != nil {
						return err
					}
				}
			}

			result.Success = true
			return nil
		})
		if rowErr != nil {
			result.Success = false
			result.FailureReason = batchEditFailureReason(rowErr)
			_ = s.batches.CreateResult(ctx, &result)
			results = append(results, result)
			failureCount++
			continue
		}

		_ = s.batches.CreateResult(ctx, &result)
		results = append(results, result)
		successCount++
	}

	job.SuccessCount = successCount
	job.FailureCount = failureCount
	_ = s.audit.Log(ctx, "item.batch_edit", "batch_edit_job", job.ID, createdBy, map[string]interface{}{
		"total":   len(edits),
		"success": successCount,
		"failure": failureCount,
	})
	return job, results, nil
}

// currentFieldValue returns the current string value of the given field on an item.
func currentFieldValue(item *domain.Item, field string) string {
	switch field {
	case "name":
		return item.Name
	case "description":
		return item.Description
	case "category":
		return item.Category
	case "brand":
		return item.Brand
	case "condition":
		return string(item.Condition)
	case "billing_model":
		return string(item.BillingModel)
	case "status":
		return string(item.Status)
	case "unit_price":
		return strconv.FormatFloat(item.UnitPrice, 'f', -1, 64)
	case "refundable_deposit":
		return strconv.FormatFloat(item.RefundableDeposit, 'f', -1, 64)
	case "quantity":
		return strconv.Itoa(item.Quantity)
	default:
		return ""
	}
}

// applyFieldChange applies a string value to the named field of an item.
// Returns an error if the field name is not recognised or the value is invalid.
func applyFieldChange(item *domain.Item, field, value string) error {
	switch field {
	case "name":
		item.Name = value
	case "description":
		item.Description = value
	case "category":
		item.Category = value
	case "brand":
		item.Brand = value
	case "condition":
		condition := domain.ItemCondition(value)
		if !condition.IsValid() {
			return &domain.ErrValidation{Field: "condition", Message: "must be one of new, open_box, or used"}
		}
		item.Condition = condition
	case "billing_model":
		billingModel := domain.BillingModel(value)
		if !billingModel.IsValid() {
			return &domain.ErrValidation{Field: "billing_model", Message: "must be one of one_time or monthly_rental"}
		}
		item.BillingModel = billingModel
	case "status":
		status := domain.ItemStatus(value)
		if !status.IsValid() {
			return &domain.ErrValidation{Field: "status", Message: "must be one of draft, published, or unpublished"}
		}
		item.Status = status
	case "unit_price":
		price, err := strconv.ParseFloat(value, 64)
		if err != nil || price < 0 {
			return &domain.ErrValidation{Field: "unit_price", Message: "must be a non-negative number"}
		}
		item.UnitPrice = price
	case "refundable_deposit":
		deposit, err := strconv.ParseFloat(value, 64)
		if err != nil || deposit < 0 {
			return &domain.ErrValidation{Field: "refundable_deposit", Message: "must be a non-negative number"}
		}
		item.RefundableDeposit = deposit
	case "quantity":
		qty, err := strconv.Atoi(value)
		if err != nil || qty < 0 {
			return &domain.ErrValidation{Field: "quantity", Message: "must be a non-negative integer"}
		}
		item.Quantity = qty
	default:
		return &domain.ErrValidation{Field: field, Message: "unknown field"}
	}
	return nil
}

func validateBatchEditState(item *domain.Item, availability []domain.AvailabilityWindow, blackouts []domain.BlackoutWindow) error {
	if err := validateAvailabilityWindows(availability); err != nil {
		return err
	}
	if err := validateBlackoutWindows(blackouts); err != nil {
		return err
	}

	if errs := item.Validate(); len(errs) > 0 {
		return &domain.ErrValidation{Field: errs[0].Field, Message: errs[0].Message}
	}

	if item.Status == domain.ItemStatusPublished {
		if blocked := domain.ValidateItemForPublish(*item, availability, blackouts); blocked != nil {
			return blocked
		}
	}

	return nil
}

func validateAvailabilityWindows(windows []domain.AvailabilityWindow) error {
	for _, window := range windows {
		if !window.EndTime.After(window.StartTime) {
			return &domain.ErrValidation{Field: "availability_windows", Message: "availability window end_time must be after start_time"}
		}
	}
	return nil
}

func validateBlackoutWindows(windows []domain.BlackoutWindow) error {
	for _, window := range windows {
		if !window.EndTime.After(window.StartTime) {
			return &domain.ErrValidation{Field: "blackout_windows", Message: "blackout window end_time must be after start_time"}
		}
	}
	return nil
}

func serializeAvailabilityWindows(windows []domain.AvailabilityWindow) string {
	if len(windows) == 0 {
		return "[]"
	}

	parts := make([]string, 0, len(windows))
	for _, window := range windows {
		parts = append(parts, window.StartTime.UTC().Format(time.RFC3339)+"|"+window.EndTime.UTC().Format(time.RFC3339))
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func cloneAvailabilityWindows(windows []domain.AvailabilityWindow) []domain.AvailabilityWindow {
	if len(windows) == 0 {
		return []domain.AvailabilityWindow{}
	}
	cloned := make([]domain.AvailabilityWindow, len(windows))
	copy(cloned, windows)
	return cloned
}

func batchEditFailureReason(err error) string {
	var validationErr *domain.ErrValidation
	if errors.As(err, &validationErr) {
		return validationErr.Message
	}

	var publishErr *domain.ErrPublishBlocked
	if errors.As(err, &publishErr) {
		return strings.Join(publishErr.Reasons, "; ")
	}

	var conflictErr *domain.ErrConflict
	if errors.As(err, &conflictErr) {
		if conflictErr.Message != "" {
			return conflictErr.Message
		}
	}

	if errors.Is(err, domain.ErrNotFound) {
		return "item not found"
	}

	return err.Error()
}
