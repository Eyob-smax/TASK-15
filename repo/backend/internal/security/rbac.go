package security

import (
	"github.com/labstack/echo/v4"

	"fitcommerce/internal/domain"
)

// contextKey is a package-local type used as context keys to avoid collisions
// with other packages that also store values in Echo's context.
type contextKey string

// UserContextKey is the key under which the authenticated user is stored in
// the Echo request context after successful session validation.
const UserContextKey contextKey = "user"

// Permission action constants used by RequireRole middleware and service-layer
// authorization checks.
const (
	ActionManageUsers       = "manage_users"
	ActionViewAudit         = "view_audit"
	ActionManageBackups     = "manage_backups"
	ActionManageBiometric   = "manage_biometric"
	ActionManageCatalog     = "manage_catalog"
	ActionViewCatalog       = "view_catalog"
	ActionCreateCampaign    = "create_campaign"
	ActionManageInventory   = "manage_inventory"
	ActionManageProcurement = "manage_procurement"
	ActionViewProcurement   = "view_procurement"
	ActionManageOrders      = "manage_orders"
	ActionViewOrders        = "view_orders"
	ActionCreateOrder       = "create_order"
	ActionJoinCampaign      = "join_campaign"
	ActionViewDashboard     = "view_dashboard"
	ActionManageReports     = "manage_reports"
	ActionViewReports       = "view_reports"
	// Dedicated permissions for personnel/location data — intentionally distinct from
	// ViewCatalog so that Member/Coach roles cannot access operational personnel data.
	ActionViewLocations = "view_locations"
	ActionViewCoaches   = "view_coaches"
	ActionViewMembers   = "view_members"
)

// rolePermissions maps each role to the set of actions it is permitted to perform.
var rolePermissions = map[domain.UserRole]map[string]bool{
	domain.UserRoleAdministrator: {
		ActionManageUsers:       true,
		ActionViewAudit:         true,
		ActionManageBackups:     true,
		ActionManageBiometric:   true,
		ActionManageCatalog:     true,
		ActionViewCatalog:       true,
		ActionCreateCampaign:    true,
		ActionManageInventory:   true,
		ActionManageProcurement: true,
		ActionViewProcurement:   true,
		ActionManageOrders:      true,
		ActionViewOrders:        true,
		ActionCreateOrder:       true,
		ActionViewDashboard:     true,
		ActionManageReports:     true,
		ActionViewReports:       true,
		ActionViewLocations:     true,
		ActionViewCoaches:       true,
		ActionViewMembers:       true,
	},
	domain.UserRoleOperationsManager: {
		ActionManageCatalog:     true,
		ActionViewCatalog:       true,
		ActionCreateCampaign:    true,
		ActionManageInventory:   true,
		ActionManageProcurement: true,
		ActionViewProcurement:   true,
		ActionManageOrders:      true,
		ActionViewOrders:        true,
		ActionCreateOrder:       true,
		ActionViewDashboard:     true,
		ActionManageReports:     true,
		ActionViewReports:       true,
		ActionViewLocations:     true,
		ActionViewCoaches:       true,
		ActionViewMembers:       true,
	},
	domain.UserRoleProcurementSpecialist: {
		ActionViewCatalog:       true,
		ActionManageProcurement: true,
		ActionViewProcurement:   true,
		ActionViewOrders:        true,
		ActionCreateOrder:       true,
		ActionViewDashboard:     true,
		ActionViewReports:       true,
		ActionViewLocations:     true,
	},
	domain.UserRoleCoach: {
		ActionViewCatalog:   true,
		ActionViewOrders:    true,
		ActionCreateOrder:   true,
		ActionViewDashboard: true,
		ActionViewReports:   true,
		ActionViewLocations: true,
	},
	domain.UserRoleMember: {
		ActionViewCatalog:    true,
		ActionCreateCampaign: true,
		ActionViewOrders:     true,
		ActionCreateOrder:    true,
		ActionJoinCampaign:   true,
	},
}

// HasPermission reports whether the given role is permitted to perform action.
func HasPermission(role domain.UserRole, action string) bool {
	perms, ok := rolePermissions[role]
	if !ok {
		return false
	}
	return perms[action]
}

// GetUserFromContext retrieves the authenticated user stored in the Echo context
// by AuthMiddleware. Returns (user, true) on success, (nil, false) if absent.
func GetUserFromContext(c echo.Context) (*domain.User, bool) {
	val := c.Get(string(UserContextKey))
	if val == nil {
		return nil, false
	}
	user, ok := val.(*domain.User)
	return user, ok
}

// SetUserInContext stores the authenticated user in the Echo request context
// so downstream handlers and middleware can retrieve it via GetUserFromContext.
func SetUserInContext(c echo.Context, user *domain.User) {
	c.Set(string(UserContextKey), user)
}
