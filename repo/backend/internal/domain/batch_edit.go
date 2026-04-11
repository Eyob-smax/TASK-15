package domain

import (
	"time"

	"github.com/google/uuid"
)

// BatchEditJob represents a batch edit operation applied to multiple items.
type BatchEditJob struct {
	ID           uuid.UUID
	CreatedBy    uuid.UUID
	CreatedAt    time.Time
	TotalRows    int
	SuccessCount int
	FailureCount int
}

// BatchEditResult records the outcome of a single row within a batch edit job.
type BatchEditResult struct {
	ID            uuid.UUID
	BatchID       uuid.UUID
	ItemID        uuid.UUID
	Field         string
	OldValue      string
	NewValue      string
	Success       bool
	FailureReason string
}
