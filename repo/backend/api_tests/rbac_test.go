package api_tests

import (
	"net/http"
	"testing"
)

func TestRBAC_AdminUsersRouteRequiresAuthentication(t *testing.T) {
	app := newIntegrationApp(t)

	rec := app.get(t, "/api/v1/admin/users", nil)
	requireStatus(t, rec, http.StatusUnauthorized)
}

func TestRBAC_MemberCannotAccessAdminUsersRoute(t *testing.T) {
	app := newIntegrationApp(t)
	member := app.seedUser(t, "member", nil)

	rec := app.get(t, "/api/v1/admin/users", app.login(t, member))
	requireStatus(t, rec, http.StatusForbidden)

	errBody := decodeError(t, rec)
	if errBody.Error.Code != "FORBIDDEN" {
		t.Fatalf("expected FORBIDDEN, got %s", errBody.Error.Code)
	}
}

func TestRBAC_AdminCanAccessAdminUsersRoute(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)

	rec := app.get(t, "/api/v1/admin/users", app.login(t, admin))
	requireStatus(t, rec, http.StatusOK)

	users, pagination := decodePaginated[map[string]any](t, rec)
	if len(users) == 0 {
		t.Fatal("expected at least one user in admin list")
	}
	if pagination.TotalCount == 0 {
		t.Fatal("expected pagination total_count to be populated")
	}
}
