package domain

import (
	"time"

	"github.com/google/uuid"
)

// BackupRun represents a single database backup execution record.
type BackupRun struct {
	ID                uuid.UUID
	ArchivePath       string
	Checksum          string
	ChecksumAlgorithm string // "sha256"
	EncryptionKeyRef  string
	Status            BackupStatus
	FileSize          int64
	StartedAt         time.Time
	CompletedAt       *time.Time
}
