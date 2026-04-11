package api_tests

import (
	"net/http"
	"testing"
)

func TestAuth_LoginSessionLogoutFlow(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)

	loginRec := app.post(t, "/api/v1/auth/login", map[string]string{
		"email":    admin.Email,
		"password": admin.Password,
	}, nil)
	requireStatus(t, loginRec, http.StatusOK)

	cookies := responseCookies(loginRec)
	if len(cookies) == 0 {
		t.Fatal("expected login to set at least one cookie")
	}

	loginBody := decodeSuccess[map[string]any](t, loginRec)
	if loginBody["user"] == nil {
		t.Fatal("expected login response to include user payload")
	}
	if loginBody["session"] == nil {
		t.Fatal("expected login response to include session payload")
	}
	sessionBody, ok := loginBody["session"].(map[string]any)
	if !ok {
		t.Fatalf("expected session payload object, got %#v", loginBody["session"])
	}
	if _, hasToken := sessionBody["token"]; hasToken {
		t.Fatal("did not expect login response to include session token in body")
	}

	sessionRec := app.get(t, "/api/v1/auth/session", cookies)
	requireStatus(t, sessionRec, http.StatusOK)

	logoutRec := app.post(t, "/api/v1/auth/logout", map[string]string{}, cookies)
	requireStatus(t, logoutRec, http.StatusOK)
	logoutBody := decodeSuccess[map[string]any](t, logoutRec)
	if logoutBody["message"] != "logged out" {
		t.Fatalf("expected logout message, got %#v", logoutBody["message"])
	}

	postLogoutSession := app.get(t, "/api/v1/auth/session", cookies)
	requireStatus(t, postLogoutSession, http.StatusUnauthorized)
}

func TestAuth_SessionEndpointRequiresAuthentication(t *testing.T) {
	app := newIntegrationApp(t)

	rec := app.get(t, "/api/v1/auth/session", nil)
	requireStatus(t, rec, http.StatusUnauthorized)

	errBody := decodeError(t, rec)
	if errBody.Error.Code != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got %s", errBody.Error.Code)
	}
}
