package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// SessionCookieName is the name of the HTTP cookie used to transmit the
// session token between the client and server.
const SessionCookieName = "fitcommerce_session"

// GenerateSessionToken creates a cryptographically random 32-byte token,
// returned as a base64url-encoded string safe for use in HTTP cookies.
func GenerateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating session token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// SetSessionCookie writes the session cookie onto the Echo response. The cookie
// is HttpOnly and SameSite=Strict. Secure is set only when the server is running
// with TLS (indicated by the secure parameter). MaxAge is derived from the
// session's absolute expiry.
func SetSessionCookie(c echo.Context, token string, absoluteExpiry time.Time, secure bool) {
	maxAge := int(time.Until(absoluteExpiry).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   secure,
		MaxAge:   maxAge,
	}
	c.SetCookie(cookie)
}

// ClearSessionCookie removes the session cookie by setting MaxAge to -1.
func ClearSessionCookie(c echo.Context) {
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	}
	c.SetCookie(cookie)
}

// ExtractSessionToken reads the session cookie from the incoming request.
// Returns an empty string if the cookie is absent or malformed.
func ExtractSessionToken(c echo.Context) string {
	cookie, err := c.Cookie(SessionCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}
