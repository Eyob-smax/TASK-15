package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ReportDefinition describes a configurable report type with role-based access.
type ReportDefinition struct {
	ID           uuid.UUID
	Name         string
	ReportType   string
	Description  string
	AllowedRoles []UserRole
	Filters      map[string]interface{}
	CreatedAt    time.Time
}

// ExportJob tracks the asynchronous generation of a report export file.
type ExportJob struct {
	ID          uuid.UUID
	ReportID    uuid.UUID
	Format      ExportFormat
	Filename    string
	Status      ExportStatus
	FilePath    string
	CreatedBy   uuid.UUID
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// GenerateExportFilename produces a standardized filename for a report export
// in the format "{reportType}_{YYYYMMDD_HHmmss}.{csv|pdf}".
func GenerateExportFilename(reportType string, format ExportFormat, now time.Time) string {
	timestamp := now.UTC().Format("20060102_150405")
	return fmt.Sprintf("%s_%s.%s", reportType, timestamp, string(format))
}
