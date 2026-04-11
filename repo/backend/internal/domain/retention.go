package domain

import (
	"time"

	"github.com/google/uuid"
)

// Retention period constants for different entity categories.
const (
	RetentionFinancialYears = 7
	RetentionAccessLogYears = 2
	RetentionFinancialDays  = RetentionFinancialYears * 365
	RetentionAccessLogDays  = RetentionAccessLogYears * 365
)

// RetentionPolicy defines how long a particular entity type must be retained.
type RetentionPolicy struct {
	ID             uuid.UUID
	EntityType     string
	RetentionDays  int
	Description    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// IsWithinRetention returns true if the entity's creation time is still within
// the retention window based on the given number of retention days from now.
func IsWithinRetention(createdAt time.Time, retentionDays int, now time.Time) bool {
	retentionEnd := createdAt.AddDate(0, 0, retentionDays)
	return now.Before(retentionEnd)
}
