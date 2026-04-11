package domain_test

import (
	"strings"
	"testing"
	"time"

	"fitcommerce/internal/domain"
)

func TestGenerateExportFilename_Format(t *testing.T) {
	fixedTime := time.Date(2026, 4, 10, 14, 30, 0, 0, time.UTC)

	filename := domain.GenerateExportFilename("items_summary", domain.ExportFormatCSV, fixedTime)

	if filename != "items_summary_20260410_143000.csv" {
		t.Errorf("unexpected filename: %q", filename)
	}
}

func TestGenerateExportFilename_PDFExtension(t *testing.T) {
	fixedTime := time.Date(2026, 4, 10, 9, 5, 3, 0, time.UTC)

	filename := domain.GenerateExportFilename("orders_summary", domain.ExportFormatPDF, fixedTime)

	if !strings.HasSuffix(filename, ".pdf") {
		t.Errorf("expected .pdf extension, got %q", filename)
	}
}

func TestGenerateExportFilename_ContainsReportType(t *testing.T) {
	fixedTime := time.Now()

	for _, reportType := range []string{"items_summary", "orders_summary", "procurement_summary", "inventory_snapshot", "landed_cost_rollup"} {
		filename := domain.GenerateExportFilename(reportType, domain.ExportFormatCSV, fixedTime)
		if !strings.HasPrefix(filename, reportType+"_") {
			t.Errorf("expected filename to start with %q, got %q", reportType+"_", filename)
		}
	}
}

func TestGenerateExportFilename_TimestampIsUTC(t *testing.T) {
	// Create a time in a non-UTC zone to verify UTC normalization.
	loc, _ := time.LoadLocation("America/New_York")
	eastTime := time.Date(2026, 4, 10, 14, 0, 0, 0, loc) // 14:00 EST = 19:00 UTC (approximate)

	filename := domain.GenerateExportFilename("items_summary", domain.ExportFormatCSV, eastTime)

	// The UTC timestamp (18:00 or 19:00 depending on DST) must not equal the local time (14:00).
	if strings.Contains(filename, "_20260410_140000") {
		t.Errorf("expected UTC timestamp, got local time in filename: %q", filename)
	}
}

func TestGenerateExportFilename_DifferentTimesProduceDifferentFilenames(t *testing.T) {
	t1 := time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 4, 10, 11, 0, 0, 0, time.UTC)

	f1 := domain.GenerateExportFilename("orders_summary", domain.ExportFormatCSV, t1)
	f2 := domain.GenerateExportFilename("orders_summary", domain.ExportFormatCSV, t2)

	if f1 == f2 {
		t.Errorf("different times should produce different filenames: %q == %q", f1, f2)
	}
}
