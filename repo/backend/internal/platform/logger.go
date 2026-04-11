package platform

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// NewLogger creates a structured JSON logger writing to stdout at the specified level.
// Supported levels: debug, info, warn, error. Defaults to info if unrecognized.
func NewLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: false,
	})

	return slog.New(handler)
}

// LogLoginAttempt logs a login attempt. The password is never logged.
func LogLoginAttempt(logger *slog.Logger, email string, success bool, reason string) {
	attrs := []any{
		"email_masked", maskEmail(email),
		"success", success,
	}
	if reason != "" {
		attrs = append(attrs, "reason", reason)
	}
	if success {
		logger.Info("login attempt", attrs...)
	} else {
		logger.Warn("login attempt", attrs...)
	}
}

// LogLockout logs when an account is locked after repeated failed login attempts.
func LogLockout(logger *slog.Logger, email string, lockedUntil time.Time) {
	logger.Warn("account locked",
		"email_masked", maskEmail(email),
		"locked_until", lockedUntil.UTC().Format(time.RFC3339),
	)
}

// LogSessionExpired logs when a session is invalidated due to timeout.
func LogSessionExpired(logger *slog.Logger, userID uuid.UUID, reason string) {
	logger.Info("session expired",
		"user_id", userID.String(),
		"reason", reason,
	)
}

// LogAuditEvent logs that an audit event was written to the tamper-evident log.
func LogAuditEvent(logger *slog.Logger, eventType, entityType string, entityID uuid.UUID) {
	logger.Info("audit event recorded",
		"event_type", eventType,
		"entity_type", entityType,
		"entity_id", entityID.String(),
	)
}

// LoggerMiddleware returns an Echo middleware function that logs each request
// with method, path, status, latency, and request ID using the provided structured logger.
func LoggerMiddleware(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()
			latency := time.Since(start)

			requestID := req.Header.Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = res.Header().Get(echo.HeaderXRequestID)
			}

			logger.Info("request",
				"method", req.Method,
				"path", req.URL.Path,
				"status", res.Status,
				"latency_ms", latency.Milliseconds(),
				"request_id", requestID,
			)

			return nil
		}
	}
}

func maskEmail(email string) string {
	at := strings.Index(email, "@")
	if at <= 0 {
		return "***@***"
	}
	return email[:1] + "***" + email[at:]
}
