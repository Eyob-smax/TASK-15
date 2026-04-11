package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"fitcommerce/internal/domain"
)

// ErrorResponse is the top-level error envelope returned by all API endpoints.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody contains the error code, human-readable message, and optional details.
type ErrorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// ErrorDetail provides field-level error information for validation failures.
type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// NewErrorResponse creates a new ErrorResponse with the given code, message,
// and optional detail entries.
func NewErrorResponse(code, message string, details ...ErrorDetail) ErrorResponse {
	return ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// HandleDomainError maps domain-level errors to appropriate HTTP status codes
// and returns a structured error envelope to the client.
func HandleDomainError(c echo.Context, err error) error {
	if err == nil {
		return nil
	}

	// Sentinel errors
	if errors.Is(err, domain.ErrNotFound) {
		return c.JSON(http.StatusNotFound, NewErrorResponse(
			"NOT_FOUND",
			err.Error(),
		))
	}
	if errors.Is(err, domain.ErrUnauthorized) {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse(
			"UNAUTHORIZED",
			err.Error(),
		))
	}
	if errors.Is(err, domain.ErrForbidden) {
		return c.JSON(http.StatusForbidden, NewErrorResponse(
			"FORBIDDEN",
			err.Error(),
		))
	}

	// Struct errors
	var validationErr *domain.ErrValidation
	if errors.As(err, &validationErr) {
		details := []ErrorDetail{
			{Field: validationErr.Field, Message: validationErr.Message},
		}
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse(
			"VALIDATION_ERROR",
			"validation failed",
			details...,
		))
	}

	var conflictErr *domain.ErrConflict
	if errors.As(err, &conflictErr) {
		return c.JSON(http.StatusConflict, NewErrorResponse(
			"CONFLICT",
			conflictErr.Error(),
		))
	}

	var lockedErr *domain.ErrAccountLocked
	if errors.As(err, &lockedErr) {
		details := []ErrorDetail{}
		if lockedErr.LockedUntil != nil {
			details = append(details, ErrorDetail{Field: "locked_until", Message: lockedErr.LockedUntil.UTC().Format("2006-01-02T15:04:05Z07:00")})
		}
		return c.JSON(http.StatusLocked, NewErrorResponse(
			"ACCOUNT_LOCKED",
			lockedErr.Error(),
			details...,
		))
	}

	var captchaErr *domain.ErrCaptchaRequired
	if errors.As(err, &captchaErr) {
		return c.JSON(http.StatusForbidden, NewErrorResponse(
			"CAPTCHA_REQUIRED",
			captchaErr.Error(),
			ErrorDetail{Field: "challenge_id", Message: captchaErr.ChallengeID},
			ErrorDetail{Field: "challenge_data", Message: captchaErr.ChallengeData},
		))
	}

	var transitionErr *domain.ErrInvalidTransition
	if errors.As(err, &transitionErr) {
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse(
			"INVALID_TRANSITION",
			transitionErr.Error(),
		))
	}

	var publishErr *domain.ErrPublishBlocked
	if errors.As(err, &publishErr) {
		var details []ErrorDetail
		for _, reason := range publishErr.Reasons {
			details = append(details, ErrorDetail{Message: reason})
		}
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse(
			"PUBLISH_BLOCKED",
			"item cannot be published",
			details...,
		))
	}

	var batchErr *domain.ErrBatchEditPartialFailure
	if errors.As(err, &batchErr) {
		var details []ErrorDetail
		for _, row := range batchErr.FailedRows {
			details = append(details, ErrorDetail{
				Field:   row.ItemID + "." + row.Field,
				Message: row.Reason,
			})
		}
		return c.JSON(http.StatusMultiStatus, NewErrorResponse(
			"BATCH_PARTIAL_FAILURE",
			batchErr.Error(),
			details...,
		))
	}

	var varianceErr *domain.ErrVarianceUnresolved
	if errors.As(err, &varianceErr) {
		return c.JSON(http.StatusUnprocessableEntity, NewErrorResponse(
			"VARIANCE_UNRESOLVED",
			varianceErr.Error(),
		))
	}

	var retentionErr *domain.ErrRetentionViolation
	if errors.As(err, &retentionErr) {
		return c.JSON(http.StatusForbidden, NewErrorResponse(
			"RETENTION_VIOLATION",
			retentionErr.Error(),
		))
	}

	// Fallback: internal server error
	return c.JSON(http.StatusInternalServerError, NewErrorResponse(
		"INTERNAL_ERROR",
		"an unexpected error occurred",
	))
}
