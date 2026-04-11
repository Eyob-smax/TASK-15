package http

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fitcommerce/internal/application"
	"fitcommerce/internal/security"
)

// NewAuthMiddleware returns middleware that validates the session cookie on each
// request. On success it stores the authenticated user in the Echo context via
// SetUserInContext. Returns 401 if the token is absent or invalid.
func NewAuthMiddleware(authSvc application.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := security.ExtractSessionToken(c)
			if token == "" {
				return c.JSON(http.StatusUnauthorized, NewErrorResponse(
					"UNAUTHORIZED", "authentication required",
				))
			}

			_, user, err := authSvc.ValidateSession(c.Request().Context(), token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, NewErrorResponse(
					"UNAUTHORIZED", "session expired or invalid",
				))
			}

			security.SetUserInContext(c, user)
			return next(c)
		}
	}
}

// NewRequireRole returns middleware that enforces that the authenticated user has
// at least one of the specified permission actions. Requires NewAuthMiddleware to
// have already run (user must be in context). Returns 403 if no action matches.
func NewRequireRole(actions ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := security.GetUserFromContext(c)
			if !ok {
				return c.JSON(http.StatusUnauthorized, NewErrorResponse(
					"UNAUTHORIZED", "authentication required",
				))
			}

			for _, action := range actions {
				if security.HasPermission(user.Role, action) {
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, NewErrorResponse(
				"FORBIDDEN", "insufficient permissions for this action",
			))
		}
	}
}

// RequestIDMiddleware generates a unique request ID for each incoming request
// and sets it on both the request and response headers.
func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqID := c.Request().Header.Get(echo.HeaderXRequestID)
			if reqID == "" {
				reqID = uuid.New().String()
			}
			c.Request().Header.Set(echo.HeaderXRequestID, reqID)
			c.Response().Header().Set(echo.HeaderXRequestID, reqID)
			return next(c)
		}
	}
}

// RecoverMiddleware catches panics in downstream handlers and converts them
// to a 500 Internal Server Error response, logging the panic details.
func RecoverMiddleware(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("panic recovered",
						"panic", r,
						"method", c.Request().Method,
						"path", c.Request().URL.Path,
					)
					_ = c.JSON(http.StatusInternalServerError, NewErrorResponse(
						"INTERNAL_ERROR", "an unexpected error occurred",
					))
				}
			}()
			return next(c)
		}
	}
}
