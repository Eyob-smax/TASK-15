package security_test

import (
	"testing"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/security"
)

// permMatrix is the authoritative permission matrix tested exhaustively.
var permMatrix = []struct {
	role   domain.UserRole
	action string
	allow  bool
}{
	// Admin — all actions
	{domain.UserRoleAdministrator, security.ActionManageUsers, true},
	{domain.UserRoleAdministrator, security.ActionViewAudit, true},
	{domain.UserRoleAdministrator, security.ActionManageBackups, true},
	{domain.UserRoleAdministrator, security.ActionManageBiometric, true},
	{domain.UserRoleAdministrator, security.ActionManageCatalog, true},
	{domain.UserRoleAdministrator, security.ActionViewCatalog, true},
	{domain.UserRoleAdministrator, security.ActionCreateCampaign, true},
	{domain.UserRoleAdministrator, security.ActionManageInventory, true},
	{domain.UserRoleAdministrator, security.ActionManageProcurement, true},
	{domain.UserRoleAdministrator, security.ActionViewProcurement, true},
	{domain.UserRoleAdministrator, security.ActionManageOrders, true},
	{domain.UserRoleAdministrator, security.ActionViewOrders, true},
	{domain.UserRoleAdministrator, security.ActionJoinCampaign, false},
	{domain.UserRoleAdministrator, security.ActionViewDashboard, true},
	{domain.UserRoleAdministrator, security.ActionManageReports, true},
	{domain.UserRoleAdministrator, security.ActionViewReports, true},

	// OperationsManager
	{domain.UserRoleOperationsManager, security.ActionManageUsers, false},
	{domain.UserRoleOperationsManager, security.ActionViewAudit, false},
	{domain.UserRoleOperationsManager, security.ActionManageBackups, false},
	{domain.UserRoleOperationsManager, security.ActionManageBiometric, false},
	{domain.UserRoleOperationsManager, security.ActionManageCatalog, true},
	{domain.UserRoleOperationsManager, security.ActionViewCatalog, true},
	{domain.UserRoleOperationsManager, security.ActionCreateCampaign, true},
	{domain.UserRoleOperationsManager, security.ActionManageInventory, true},
	{domain.UserRoleOperationsManager, security.ActionManageProcurement, true},
	{domain.UserRoleOperationsManager, security.ActionViewProcurement, true},
	{domain.UserRoleOperationsManager, security.ActionManageOrders, true},
	{domain.UserRoleOperationsManager, security.ActionViewOrders, true},
	{domain.UserRoleOperationsManager, security.ActionJoinCampaign, false},
	{domain.UserRoleOperationsManager, security.ActionViewDashboard, true},
	{domain.UserRoleOperationsManager, security.ActionManageReports, true},
	{domain.UserRoleOperationsManager, security.ActionViewReports, true},

	// ProcurementSpecialist
	{domain.UserRoleProcurementSpecialist, security.ActionManageUsers, false},
	{domain.UserRoleProcurementSpecialist, security.ActionViewAudit, false},
	{domain.UserRoleProcurementSpecialist, security.ActionManageBackups, false},
	{domain.UserRoleProcurementSpecialist, security.ActionManageBiometric, false},
	{domain.UserRoleProcurementSpecialist, security.ActionManageCatalog, false},
	{domain.UserRoleProcurementSpecialist, security.ActionViewCatalog, true},
	{domain.UserRoleProcurementSpecialist, security.ActionCreateCampaign, false},
	{domain.UserRoleProcurementSpecialist, security.ActionManageInventory, false},
	{domain.UserRoleProcurementSpecialist, security.ActionManageProcurement, true},
	{domain.UserRoleProcurementSpecialist, security.ActionViewProcurement, true},
	{domain.UserRoleProcurementSpecialist, security.ActionManageOrders, false},
	{domain.UserRoleProcurementSpecialist, security.ActionViewOrders, true},
	{domain.UserRoleProcurementSpecialist, security.ActionJoinCampaign, false},
	{domain.UserRoleProcurementSpecialist, security.ActionViewDashboard, true},
	{domain.UserRoleProcurementSpecialist, security.ActionManageReports, false},
	{domain.UserRoleProcurementSpecialist, security.ActionViewReports, true},

	// Coach
	{domain.UserRoleCoach, security.ActionManageUsers, false},
	{domain.UserRoleCoach, security.ActionViewAudit, false},
	{domain.UserRoleCoach, security.ActionManageBackups, false},
	{domain.UserRoleCoach, security.ActionManageBiometric, false},
	{domain.UserRoleCoach, security.ActionManageCatalog, false},
	{domain.UserRoleCoach, security.ActionViewCatalog, true},
	{domain.UserRoleCoach, security.ActionCreateCampaign, false},
	{domain.UserRoleCoach, security.ActionManageInventory, false},
	{domain.UserRoleCoach, security.ActionManageProcurement, false},
	{domain.UserRoleCoach, security.ActionViewProcurement, false},
	{domain.UserRoleCoach, security.ActionManageOrders, false},
	{domain.UserRoleCoach, security.ActionViewOrders, true},
	{domain.UserRoleCoach, security.ActionJoinCampaign, false},
	{domain.UserRoleCoach, security.ActionViewDashboard, true},
	{domain.UserRoleCoach, security.ActionManageReports, false},
	{domain.UserRoleCoach, security.ActionViewReports, true},

	// Member
	{domain.UserRoleMember, security.ActionManageUsers, false},
	{domain.UserRoleMember, security.ActionViewAudit, false},
	{domain.UserRoleMember, security.ActionManageBackups, false},
	{domain.UserRoleMember, security.ActionManageBiometric, false},
	{domain.UserRoleMember, security.ActionManageCatalog, false},
	{domain.UserRoleMember, security.ActionViewCatalog, true},
	{domain.UserRoleMember, security.ActionCreateCampaign, true},
	{domain.UserRoleMember, security.ActionManageInventory, false},
	{domain.UserRoleMember, security.ActionManageProcurement, false},
	{domain.UserRoleMember, security.ActionViewProcurement, false},
	{domain.UserRoleMember, security.ActionManageOrders, false},
	{domain.UserRoleMember, security.ActionViewOrders, true},
	{domain.UserRoleMember, security.ActionJoinCampaign, true},
	{domain.UserRoleMember, security.ActionViewDashboard, false},
	{domain.UserRoleMember, security.ActionManageReports, false},
	{domain.UserRoleMember, security.ActionViewReports, false},
}

func TestHasPermission_Matrix(t *testing.T) {
	for _, tc := range permMatrix {
		got := security.HasPermission(tc.role, tc.action)
		if got != tc.allow {
			t.Errorf("HasPermission(role=%s, action=%s) = %v, want %v",
				tc.role, tc.action, got, tc.allow)
		}
	}
}

func TestHasPermission_UnknownRole(t *testing.T) {
	if security.HasPermission("unknown_role", security.ActionViewCatalog) {
		t.Error("expected false for unknown role")
	}
}

func TestHasPermission_UnknownAction(t *testing.T) {
	if security.HasPermission(domain.UserRoleAdministrator, "nonexistent_action") {
		t.Error("expected false for unknown action even for admin")
	}
}
