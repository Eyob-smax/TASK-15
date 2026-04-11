package security_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"fitcommerce/internal/security"
)

func TestGenerateSessionToken_Unique(t *testing.T) {
	tokens := make(map[string]bool, 3)
	for i := 0; i < 3; i++ {
		tok, err := security.GenerateSessionToken()
		if err != nil {
			t.Fatalf("GenerateSessionToken error: %v", err)
		}
		if tok == "" {
			t.Fatal("GenerateSessionToken returned empty string")
		}
		if tokens[tok] {
			t.Fatalf("duplicate token generated: %q", tok)
		}
		tokens[tok] = true
	}
}

func TestSessionCookieName(t *testing.T) {
	if security.SessionCookieName != "fitcommerce_session" {
		t.Errorf("unexpected SessionCookieName: %q", security.SessionCookieName)
	}
}

func TestSetSessionCookie_Attributes(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expiry := time.Now().Add(12 * time.Hour)
	security.SetSessionCookie(c, "test-token", expiry, false)

	cookies := rec.Result().Cookies()
	var found *http.Cookie
	for _, ck := range cookies {
		if ck.Name == security.SessionCookieName {
			found = ck
			break
		}
	}
	if found == nil {
		t.Fatal("session cookie not set in response")
	}
	if found.Value != "test-token" {
		t.Errorf("unexpected cookie value: %q", found.Value)
	}
	if !found.HttpOnly {
		t.Error("session cookie must be HttpOnly")
	}
	if found.SameSite != http.SameSiteStrictMode {
		t.Error("session cookie must have SameSite=Strict")
	}
	if found.Secure {
		t.Error("secure should be false when TLS is disabled")
	}
}

func TestClearSessionCookie(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	security.ClearSessionCookie(c)

	cookies := rec.Result().Cookies()
	var found *http.Cookie
	for _, ck := range cookies {
		if ck.Name == security.SessionCookieName {
			found = ck
			break
		}
	}
	if found == nil {
		t.Fatal("expected clear cookie to be set")
	}
	if found.MaxAge != -1 {
		t.Errorf("expected MaxAge=-1 for cleared cookie, got %d", found.MaxAge)
	}
}

func TestExtractSessionToken_Missing(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	tok := security.ExtractSessionToken(c)
	if tok != "" {
		t.Errorf("expected empty token when cookie absent, got %q", tok)
	}
}

func TestExtractSessionToken_Present(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: security.SessionCookieName, Value: "mytoken"})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	tok := security.ExtractSessionToken(c)
	if tok != "mytoken" {
		t.Errorf("expected %q, got %q", "mytoken", tok)
	}
}
