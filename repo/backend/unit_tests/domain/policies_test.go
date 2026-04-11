package domain_test

import (
	"strings"
	"testing"
	"time"

	"fitcommerce/internal/domain"

	"github.com/google/uuid"
)

// ─── Helpers ───────────────────────────────────────────────────────────────────

func validTestItem() domain.Item {
	return domain.Item{
		ID:                uuid.New(),
		Name:              "Adjustable Dumbbell Set",
		Description:       "Cast-iron dumbbell set, 5-50 lb",
		Category:          "free_weights",
		Brand:             "IronGrip",
		Condition:         domain.ItemConditionNew,
		RefundableDeposit: 75.00,
		BillingModel:      domain.BillingModelOneTime,
		Status:            domain.ItemStatusDraft,
		Quantity:          10,
		CreatedBy:         uuid.New(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		Version:           1,
	}
}

// ─── ValidateItemForPublish ────────────────────────────────────────────────────

func TestValidateItemForPublish_ValidItem(t *testing.T) {
	item := validTestItem()
	err := domain.ValidateItemForPublish(item, nil, nil)
	if err != nil {
		t.Errorf("expected nil error for valid item, got: %v", err)
	}
}

func TestValidateItemForPublish_MissingName(t *testing.T) {
	item := validTestItem()
	item.Name = ""
	err := domain.ValidateItemForPublish(item, nil, nil)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	found := false
	for _, r := range err.Reasons {
		if strings.Contains(r, "name") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected reason mentioning 'name', got reasons: %v", err.Reasons)
	}
}

func TestValidateItemForPublish_MissingCategory(t *testing.T) {
	item := validTestItem()
	item.Category = ""
	err := domain.ValidateItemForPublish(item, nil, nil)
	if err == nil {
		t.Fatal("expected error for missing category")
	}
	found := false
	for _, r := range err.Reasons {
		if strings.Contains(r, "category") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected reason mentioning 'category', got: %v", err.Reasons)
	}
}

func TestValidateItemForPublish_MissingBrand(t *testing.T) {
	item := validTestItem()
	item.Brand = ""
	err := domain.ValidateItemForPublish(item, nil, nil)
	if err == nil {
		t.Fatal("expected error for missing brand")
	}
	found := false
	for _, r := range err.Reasons {
		if strings.Contains(r, "brand") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected reason mentioning 'brand', got: %v", err.Reasons)
	}
}

func TestValidateItemForPublish_InvalidCondition(t *testing.T) {
	item := validTestItem()
	item.Condition = domain.ItemCondition("broken")
	err := domain.ValidateItemForPublish(item, nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid condition")
	}
	found := false
	for _, r := range err.Reasons {
		if strings.Contains(r, "condition") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected reason mentioning 'condition', got: %v", err.Reasons)
	}
}

func TestValidateItemForPublish_InvalidBillingModel(t *testing.T) {
	item := validTestItem()
	item.BillingModel = domain.BillingModel("weekly")
	err := domain.ValidateItemForPublish(item, nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid billing model")
	}
	found := false
	for _, r := range err.Reasons {
		if strings.Contains(r, "billing") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected reason mentioning 'billing', got: %v", err.Reasons)
	}
}

func TestValidateItemForPublish_NegativeQuantity(t *testing.T) {
	item := validTestItem()
	item.Quantity = -5
	err := domain.ValidateItemForPublish(item, nil, nil)
	if err == nil {
		t.Fatal("expected error for negative quantity")
	}
	found := false
	for _, r := range err.Reasons {
		if strings.Contains(r, "quantity") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected reason mentioning 'quantity', got: %v", err.Reasons)
	}
}

func TestValidateItemForPublish_MultipleErrors(t *testing.T) {
	item := validTestItem()
	item.Name = ""
	item.Brand = ""
	err := domain.ValidateItemForPublish(item, nil, nil)
	if err == nil {
		t.Fatal("expected error for multiple missing fields")
	}
	if len(err.Reasons) < 2 {
		t.Errorf("expected at least 2 reasons, got %d: %v", len(err.Reasons), err.Reasons)
	}
	hasName := false
	hasBrand := false
	for _, r := range err.Reasons {
		if strings.Contains(r, "name") {
			hasName = true
		}
		if strings.Contains(r, "brand") {
			hasBrand = true
		}
	}
	if !hasName || !hasBrand {
		t.Errorf("expected reasons for both 'name' and 'brand', got: %v", err.Reasons)
	}
}

func TestValidateItemForPublish_WindowOverlap(t *testing.T) {
	item := validTestItem()
	avail := []domain.AvailabilityWindow{
		{
			ID:        uuid.New(),
			ItemID:    item.ID,
			StartTime: time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 5, 10, 18, 0, 0, 0, time.UTC),
		},
	}
	blackout := []domain.BlackoutWindow{
		{
			ID:        uuid.New(),
			ItemID:    item.ID,
			StartTime: time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 5, 7, 0, 0, 0, 0, time.UTC),
		},
	}
	err := domain.ValidateItemForPublish(item, avail, blackout)
	if err == nil {
		t.Fatal("expected error for overlapping windows")
	}
	found := false
	for _, r := range err.Reasons {
		if strings.Contains(r, "overlaps") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected overlap reason, got: %v", err.Reasons)
	}
}

func TestValidateItemForPublish_NoOverlap(t *testing.T) {
	item := validTestItem()
	avail := []domain.AvailabilityWindow{
		{
			ID:        uuid.New(),
			ItemID:    item.ID,
			StartTime: time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC),
		},
	}
	blackout := []domain.BlackoutWindow{
		{
			ID:        uuid.New(),
			ItemID:    item.ID,
			StartTime: time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC),
		},
	}
	err := domain.ValidateItemForPublish(item, avail, blackout)
	if err != nil {
		t.Errorf("expected no error for non-overlapping windows, got: %v", err)
	}
}

// ─── DetectWindowOverlap ───────────────────────────────────────────────────────

func TestDetectWindowOverlap_Overlapping(t *testing.T) {
	itemID := uuid.New()
	avail := []domain.AvailabilityWindow{
		{
			ID:        uuid.New(),
			ItemID:    itemID,
			StartTime: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
		},
	}
	blackout := []domain.BlackoutWindow{
		{
			ID:        uuid.New(),
			ItemID:    itemID,
			StartTime: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC),
		},
	}
	overlaps := domain.DetectWindowOverlap(avail, blackout)
	if len(overlaps) != 1 {
		t.Errorf("expected 1 overlap, got %d", len(overlaps))
	}
}

func TestDetectWindowOverlap_Adjacent(t *testing.T) {
	itemID := uuid.New()
	// Adjacent: availability ends exactly when blackout starts -> no overlap
	// The algorithm uses a < d and c < b (half-open intervals), so end == start means no overlap.
	avail := []domain.AvailabilityWindow{
		{
			ID:        uuid.New(),
			ItemID:    itemID,
			StartTime: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
		},
	}
	blackout := []domain.BlackoutWindow{
		{
			ID:        uuid.New(),
			ItemID:    itemID,
			StartTime: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
		},
	}
	overlaps := domain.DetectWindowOverlap(avail, blackout)
	if len(overlaps) != 0 {
		t.Errorf("expected 0 overlaps for adjacent windows, got %d: %v", len(overlaps), overlaps)
	}
}

func TestDetectWindowOverlap_NoOverlap(t *testing.T) {
	itemID := uuid.New()
	avail := []domain.AvailabilityWindow{
		{
			ID:        uuid.New(),
			ItemID:    itemID,
			StartTime: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		},
	}
	blackout := []domain.BlackoutWindow{
		{
			ID:        uuid.New(),
			ItemID:    itemID,
			StartTime: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC),
		},
	}
	overlaps := domain.DetectWindowOverlap(avail, blackout)
	if len(overlaps) != 0 {
		t.Errorf("expected 0 overlaps, got %d", len(overlaps))
	}
}

func TestDetectWindowOverlap_MultipleOverlaps(t *testing.T) {
	itemID := uuid.New()
	avail := []domain.AvailabilityWindow{
		{
			ID:        uuid.New(),
			ItemID:    itemID,
			StartTime: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:        uuid.New(),
			ItemID:    itemID,
			StartTime: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
		},
	}
	blackout := []domain.BlackoutWindow{
		{
			ID:        uuid.New(),
			ItemID:    itemID,
			StartTime: time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC),
		},
	}
	overlaps := domain.DetectWindowOverlap(avail, blackout)
	if len(overlaps) != 2 {
		t.Errorf("expected 2 overlaps, got %d: %v", len(overlaps), overlaps)
	}
}

// ─── ValidateQuantityNonNegative ───────────────────────────────────────────────

func TestValidateQuantityNonNegative_Valid(t *testing.T) {
	tests := []struct {
		name string
		qty  int
	}{
		{"zero", 0},
		{"positive", 42},
		{"large positive", 999999},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := domain.ValidateQuantityNonNegative(tc.qty); err != nil {
				t.Errorf("expected no error for qty=%d, got: %v", tc.qty, err)
			}
		})
	}
}

func TestValidateQuantityNonNegative_Negative(t *testing.T) {
	err := domain.ValidateQuantityNonNegative(-1)
	if err == nil {
		t.Fatal("expected error for negative quantity")
	}
	if !strings.Contains(err.Error(), "quantity") {
		t.Errorf("expected error mentioning 'quantity', got: %v", err)
	}
}

// ─── ApplyDepositDefault ───────────────────────────────────────────────────────

func TestApplyDepositDefault_ZeroValue(t *testing.T) {
	result := domain.ApplyDepositDefault(0)
	if result != 50.00 {
		t.Errorf("expected 50.00 for zero deposit, got %f", result)
	}
}

func TestApplyDepositDefault_CustomValue(t *testing.T) {
	result := domain.ApplyDepositDefault(99.99)
	if result != 99.99 {
		t.Errorf("expected 99.99 for custom deposit, got %f", result)
	}
}

// ─── Item.ApplyDepositDefault ──────────────────────────────────────────────────

func TestItem_ApplyDepositDefault(t *testing.T) {
	t.Run("sets default on zero deposit", func(t *testing.T) {
		item := validTestItem()
		item.RefundableDeposit = 0
		item.ApplyDepositDefault()
		if item.RefundableDeposit != domain.DefaultRefundableDeposit {
			t.Errorf("expected deposit %f, got %f", domain.DefaultRefundableDeposit, item.RefundableDeposit)
		}
	})

	t.Run("preserves non-zero deposit", func(t *testing.T) {
		item := validTestItem()
		item.RefundableDeposit = 125.00
		item.ApplyDepositDefault()
		if item.RefundableDeposit != 125.00 {
			t.Errorf("expected deposit 125.00, got %f", item.RefundableDeposit)
		}
	})
}

// ─── Item.Validate ─────────────────────────────────────────────────────────────

func TestItem_Validate_ValidItem(t *testing.T) {
	item := validTestItem()
	errs := item.Validate()
	if errs != nil {
		t.Errorf("expected no validation errors for valid item, got %d: %v", len(errs), errs)
	}
}

func TestItem_Validate_InvalidItem(t *testing.T) {
	item := domain.Item{
		// All required fields missing / invalid
		Condition:    domain.ItemCondition("bad"),
		BillingModel: domain.BillingModel("bad"),
		Quantity:     -1,
		RefundableDeposit: -10,
	}
	errs := item.Validate()
	if errs == nil {
		t.Fatal("expected validation errors for invalid item")
	}
	// Should have errors for: name, category, brand, condition, billing_model, quantity, refundable_deposit
	if len(errs) < 5 {
		t.Errorf("expected at least 5 validation errors, got %d: %v", len(errs), errs)
	}
}
