package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fitcommerce/internal/application"
	"fitcommerce/internal/http/dto"
	"fitcommerce/internal/platform"
	"fitcommerce/internal/security"
)

// AuthHandler handles HTTP requests for the authentication and session endpoints.
type AuthHandler struct {
	authSvc application.AuthService
	cfg     *platform.Config
}

// NewAuthHandler creates an AuthHandler with the given auth service and config.
func NewAuthHandler(authSvc application.AuthService, cfg *platform.Config) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, cfg: cfg}
}

// Login handles POST /api/v1/auth/login. On success it sets the session cookie
// and returns the authenticated user with session timing metadata. The session
// token itself is never included in the response body.
func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse(
			"BAD_REQUEST", "invalid request body",
		))
	}

	session, user, err := h.authSvc.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return HandleDomainError(c, err)
	}

	// Mark the cookie Secure if TLS is terminated locally OR if an upstream proxy
	// (e.g., nginx) indicates the external connection is HTTPS via X-Forwarded-Proto.
	secure := c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https"
	security.SetSessionCookie(c, session.Token, session.AbsoluteExpiresAt, secure)

	resp := dto.LoginResponse{
		User: dto.UserResponse{
			ID:          user.ID.String(),
			Email:       user.Email,
			Role:        string(user.Role),
			Status:      string(user.Status),
			DisplayName: user.DisplayName,
			CreatedAt:   user.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   user.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		},
		Session: dto.SessionMetaResponse{
			IdleExpiresAt:     session.IdleExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			AbsoluteExpiresAt: session.AbsoluteExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		},
	}
	if user.LocationID != nil {
		locStr := user.LocationID.String()
		resp.User.LocationID = &locStr
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: resp})
}

// Logout handles POST /api/v1/auth/logout. Invalidates the current session and
// clears the session cookie.
func (h *AuthHandler) Logout(c echo.Context) error {
	token := security.ExtractSessionToken(c)
	if token != "" {
		_ = h.authSvc.Logout(c.Request().Context(), token)
	}
	security.ClearSessionCookie(c)
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: map[string]string{"message": "logged out"}})
}

// GetSession handles GET /api/v1/auth/session. Returns the authenticated user
// and session timing for the current request. Requires the auth middleware.
func (h *AuthHandler) GetSession(c echo.Context) error {
	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse(
			"UNAUTHORIZED", "not authenticated",
		))
	}

	sess, _, err := h.authSvc.ValidateSession(c.Request().Context(), security.ExtractSessionToken(c))
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := dto.LoginResponse{
		User: dto.UserResponse{
			ID:          user.ID.String(),
			Email:       user.Email,
			Role:        string(user.Role),
			Status:      string(user.Status),
			DisplayName: user.DisplayName,
			CreatedAt:   user.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   user.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		},
		Session: dto.SessionMetaResponse{
			IdleExpiresAt:     sess.IdleExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			AbsoluteExpiresAt: sess.AbsoluteExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		},
	}
	if user.LocationID != nil {
		locStr := user.LocationID.String()
		resp.User.LocationID = &locStr
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: resp})
}

// VerifyCaptcha handles POST /api/v1/auth/captcha/verify. Accepts a challenge ID
// and submitted answer. On success the lockout state is cleared and the client
// may retry login.
func (h *AuthHandler) VerifyCaptcha(c echo.Context) error {
	var req dto.CaptchaVerifyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse(
			"BAD_REQUEST", "invalid request body",
		))
	}

	challengeID, err := uuid.Parse(req.ChallengeID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse(
			"BAD_REQUEST", "invalid challenge_id format",
		))
	}

	if err := h.authSvc.VerifyCaptcha(c.Request().Context(), challengeID, req.Answer); err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: map[string]string{"message": "captcha verified"}})
}
