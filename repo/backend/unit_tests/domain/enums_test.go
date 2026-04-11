package domain_test

import (
	"testing"

	"fitcommerce/internal/domain"
)

// ─── UserRole ──────────────────────────────────────────────────────────────────

func TestUserRole_IsValid(t *testing.T) {
	t.Run("valid roles return true", func(t *testing.T) {
		valid := []domain.UserRole{
			domain.UserRoleAdministrator,
			domain.UserRoleOperationsManager,
			domain.UserRoleProcurementSpecialist,
			domain.UserRoleCoach,
			domain.UserRoleMember,
		}
		for _, role := range valid {
			if !role.IsValid() {
				t.Errorf("expected %q to be valid", role)
			}
		}
	})

	t.Run("invalid role returns false", func(t *testing.T) {
		invalid := domain.UserRole("superadmin")
		if invalid.IsValid() {
			t.Errorf("expected %q to be invalid", invalid)
		}
	})

	t.Run("empty string is invalid", func(t *testing.T) {
		empty := domain.UserRole("")
		if empty.IsValid() {
			t.Error("expected empty string to be invalid")
		}
	})
}

func TestUserRole_IsStaff(t *testing.T) {
	staffRoles := []domain.UserRole{
		domain.UserRoleAdministrator,
		domain.UserRoleOperationsManager,
		domain.UserRoleProcurementSpecialist,
	}
	nonStaffRoles := []domain.UserRole{
		domain.UserRoleCoach,
		domain.UserRoleMember,
	}

	for _, role := range staffRoles {
		t.Run(string(role)+"_is_staff", func(t *testing.T) {
			if !role.IsStaff() {
				t.Errorf("expected %q to be staff", role)
			}
		})
	}

	for _, role := range nonStaffRoles {
		t.Run(string(role)+"_is_not_staff", func(t *testing.T) {
			if role.IsStaff() {
				t.Errorf("expected %q to NOT be staff", role)
			}
		})
	}
}

func TestUserRole_CanAccessDashboard(t *testing.T) {
	canAccess := []domain.UserRole{
		domain.UserRoleAdministrator,
		domain.UserRoleOperationsManager,
		domain.UserRoleProcurementSpecialist,
		domain.UserRoleCoach,
	}
	cannotAccess := []domain.UserRole{
		domain.UserRoleMember,
	}

	for _, role := range canAccess {
		t.Run(string(role)+"_can_access", func(t *testing.T) {
			if !role.CanAccessDashboard() {
				t.Errorf("expected %q to have dashboard access", role)
			}
		})
	}

	for _, role := range cannotAccess {
		t.Run(string(role)+"_cannot_access", func(t *testing.T) {
			if role.CanAccessDashboard() {
				t.Errorf("expected %q to NOT have dashboard access", role)
			}
		})
	}
}

// ─── ItemCondition ─────────────────────────────────────────────────────────────

func TestItemCondition_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		cond  domain.ItemCondition
		valid bool
	}{
		{"new is valid", domain.ItemConditionNew, true},
		{"open_box is valid", domain.ItemConditionOpenBox, true},
		{"used is valid", domain.ItemConditionUsed, true},
		{"refurbished is invalid", domain.ItemCondition("refurbished"), false},
		{"empty is invalid", domain.ItemCondition(""), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.cond.IsValid(); got != tc.valid {
				t.Errorf("ItemCondition(%q).IsValid() = %v, want %v", tc.cond, got, tc.valid)
			}
		})
	}
}

// ─── BillingModel ──────────────────────────────────────────────────────────────

func TestBillingModel_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		model domain.BillingModel
		valid bool
	}{
		{"one_time is valid", domain.BillingModelOneTime, true},
		{"monthly_rental is valid", domain.BillingModelMonthlyRental, true},
		{"annual is invalid", domain.BillingModel("annual"), false},
		{"empty is invalid", domain.BillingModel(""), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.model.IsValid(); got != tc.valid {
				t.Errorf("BillingModel(%q).IsValid() = %v, want %v", tc.model, got, tc.valid)
			}
		})
	}
}

// ─── ItemStatus ────────────────────────────────────────────────────────────────

func TestItemStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status domain.ItemStatus
		valid  bool
	}{
		{"draft is valid", domain.ItemStatusDraft, true},
		{"published is valid", domain.ItemStatusPublished, true},
		{"unpublished is valid", domain.ItemStatusUnpublished, true},
		{"archived is invalid", domain.ItemStatus("archived"), false},
		{"empty is invalid", domain.ItemStatus(""), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.status.IsValid(); got != tc.valid {
				t.Errorf("ItemStatus(%q).IsValid() = %v, want %v", tc.status, got, tc.valid)
			}
		})
	}
}

// ─── OrderStatus ───────────────────────────────────────────────────────────────

func TestOrderStatus_IsValid(t *testing.T) {
	allValid := []domain.OrderStatus{
		domain.OrderStatusCreated,
		domain.OrderStatusPaid,
		domain.OrderStatusCancelled,
		domain.OrderStatusRefunded,
		domain.OrderStatusAutoClosed,
	}
	for _, s := range allValid {
		t.Run(string(s)+"_is_valid", func(t *testing.T) {
			if !s.IsValid() {
				t.Errorf("expected %q to be valid", s)
			}
		})
	}

	t.Run("invalid order status", func(t *testing.T) {
		invalid := domain.OrderStatus("shipped")
		if invalid.IsValid() {
			t.Errorf("expected %q to be invalid", invalid)
		}
	})
}

func TestOrderStatus_IsTerminal(t *testing.T) {
	terminal := []domain.OrderStatus{
		domain.OrderStatusCancelled,
		domain.OrderStatusRefunded,
		domain.OrderStatusAutoClosed,
	}
	nonTerminal := []domain.OrderStatus{
		domain.OrderStatusCreated,
		domain.OrderStatusPaid,
	}

	for _, s := range terminal {
		t.Run(string(s)+"_is_terminal", func(t *testing.T) {
			if !s.IsTerminal() {
				t.Errorf("expected %q to be terminal", s)
			}
		})
	}

	for _, s := range nonTerminal {
		t.Run(string(s)+"_is_not_terminal", func(t *testing.T) {
			if s.IsTerminal() {
				t.Errorf("expected %q to NOT be terminal", s)
			}
		})
	}
}

// ─── CampaignStatus ────────────────────────────────────────────────────────────

func TestCampaignStatus_IsValid(t *testing.T) {
	allValid := []domain.CampaignStatus{
		domain.CampaignStatusActive,
		domain.CampaignStatusSucceeded,
		domain.CampaignStatusFailed,
		domain.CampaignStatusCancelled,
	}
	for _, s := range allValid {
		t.Run(string(s)+"_is_valid", func(t *testing.T) {
			if !s.IsValid() {
				t.Errorf("expected %q to be valid", s)
			}
		})
	}

	t.Run("invalid campaign status", func(t *testing.T) {
		invalid := domain.CampaignStatus("paused")
		if invalid.IsValid() {
			t.Errorf("expected %q to be invalid", invalid)
		}
	})
}

func TestCampaignStatus_IsTerminal(t *testing.T) {
	terminal := []domain.CampaignStatus{
		domain.CampaignStatusSucceeded,
		domain.CampaignStatusFailed,
		domain.CampaignStatusCancelled,
	}
	nonTerminal := []domain.CampaignStatus{
		domain.CampaignStatusActive,
	}

	for _, s := range terminal {
		t.Run(string(s)+"_is_terminal", func(t *testing.T) {
			if !s.IsTerminal() {
				t.Errorf("expected %q to be terminal", s)
			}
		})
	}

	for _, s := range nonTerminal {
		t.Run(string(s)+"_is_not_terminal", func(t *testing.T) {
			if s.IsTerminal() {
				t.Errorf("expected %q to NOT be terminal", s)
			}
		})
	}
}

// ─── POStatus ──────────────────────────────────────────────────────────────────

func TestPOStatus_IsValid(t *testing.T) {
	allValid := []domain.POStatus{
		domain.POStatusCreated,
		domain.POStatusApproved,
		domain.POStatusReceived,
		domain.POStatusReturned,
		domain.POStatusVoided,
	}
	for _, s := range allValid {
		t.Run(string(s)+"_is_valid", func(t *testing.T) {
			if !s.IsValid() {
				t.Errorf("expected %q to be valid", s)
			}
		})
	}

	t.Run("invalid PO status", func(t *testing.T) {
		invalid := domain.POStatus("shipped")
		if invalid.IsValid() {
			t.Errorf("expected %q to be invalid", invalid)
		}
	})
}

func TestPOStatus_IsTerminal(t *testing.T) {
	terminal := []domain.POStatus{
		domain.POStatusReturned,
		domain.POStatusVoided,
	}
	nonTerminal := []domain.POStatus{
		domain.POStatusCreated,
		domain.POStatusApproved,
		domain.POStatusReceived,
	}

	for _, s := range terminal {
		t.Run(string(s)+"_is_terminal", func(t *testing.T) {
			if !s.IsTerminal() {
				t.Errorf("expected %q to be terminal", s)
			}
		})
	}

	for _, s := range nonTerminal {
		t.Run(string(s)+"_is_not_terminal", func(t *testing.T) {
			if s.IsTerminal() {
				t.Errorf("expected %q to NOT be terminal", s)
			}
		})
	}
}

// ─── All* collection helpers ───────────────────────────────────────────────────

func TestAllUserRoles_ReturnsAll(t *testing.T) {
	roles := domain.AllUserRoles()
	if len(roles) != 5 {
		t.Errorf("expected 5 user roles, got %d", len(roles))
	}
}

func TestAllItemConditions_ReturnsAll(t *testing.T) {
	conditions := domain.AllItemConditions()
	if len(conditions) != 3 {
		t.Errorf("expected 3 item conditions, got %d", len(conditions))
	}
}

func TestAllBillingModels_ReturnsAll(t *testing.T) {
	models := domain.AllBillingModels()
	if len(models) != 2 {
		t.Errorf("expected 2 billing models, got %d", len(models))
	}
}
