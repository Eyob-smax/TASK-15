package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Sentinel errors for common domain conditions.
var (
	ErrNotFound     = errors.New("entity not found")
	ErrUnauthorized = errors.New("not authenticated")
	ErrForbidden    = errors.New("not authorized for this action")
)

// ErrConflict represents a version conflict or duplicate entity error.
type ErrConflict struct {
	Entity  string
	Message string
}

func (e *ErrConflict) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("conflict on %s: %s", e.Entity, e.Message)
	}
	return fmt.Sprintf("conflict on %s", e.Entity)
}

// ErrValidation represents a validation failure on a specific field.
type ErrValidation struct {
	Field   string
	Message string
}

func (e *ErrValidation) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// ValidationError is a lightweight validation error used in batch validation results.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ErrAccountLocked indicates the account is locked due to too many failed login attempts.
type ErrAccountLocked struct {
	Message     string
	LockedUntil *time.Time
}

func (e *ErrAccountLocked) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("account locked: %s", e.Message)
	}
	return "account locked due to too many failed login attempts"
}

// ErrCaptchaRequired indicates that a CAPTCHA challenge must be solved before proceeding.
type ErrCaptchaRequired struct {
	ChallengeID   string
	ChallengeData string
}

func (e *ErrCaptchaRequired) Error() string {
	return fmt.Sprintf("captcha required: challenge_id=%s", e.ChallengeID)
}

// ErrInvalidTransition represents an invalid state machine transition.
type ErrInvalidTransition struct {
	Entity string
	From   string
	To     string
}

func (e *ErrInvalidTransition) Error() string {
	return fmt.Sprintf("invalid %s transition from %s to %s", e.Entity, e.From, e.To)
}

// ErrPublishBlocked indicates that an item cannot be published due to validation failures.
type ErrPublishBlocked struct {
	Reasons []string
}

func (e *ErrPublishBlocked) Error() string {
	return fmt.Sprintf("publish blocked: %s", strings.Join(e.Reasons, "; "))
}

// BatchEditFailedRow represents a single row failure within a batch edit operation.
type BatchEditFailedRow struct {
	ItemID string
	Field  string
	Reason string
}

// ErrBatchEditPartialFailure indicates that some rows in a batch edit failed.
type ErrBatchEditPartialFailure struct {
	FailedRows []BatchEditFailedRow
}

func (e *ErrBatchEditPartialFailure) Error() string {
	return fmt.Sprintf("batch edit partial failure: %d rows failed", len(e.FailedRows))
}

// ErrVarianceUnresolved indicates that a variance record has not yet been resolved.
type ErrVarianceUnresolved struct {
	VarianceID string
}

func (e *ErrVarianceUnresolved) Error() string {
	return fmt.Sprintf("variance %s is not yet resolved", e.VarianceID)
}

// ErrRetentionViolation indicates that an entity cannot be deleted because
// it is still within the retention period.
type ErrRetentionViolation struct {
	EntityType    string
	RetentionDays int
}

func (e *ErrRetentionViolation) Error() string {
	return fmt.Sprintf("cannot delete %s: within %d-day retention period", e.EntityType, e.RetentionDays)
}
