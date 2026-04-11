package domain_test

import (
	"testing"
	"time"

	"fitcommerce/internal/domain"
)

func TestIsWithinRetention_WithinPeriod(t *testing.T) {
	// Record from 5 years ago with 7-year retention policy: within retention
	now := time.Date(2026, 4, 9, 0, 0, 0, 0, time.UTC)
	createdAt := now.AddDate(-5, 0, 0)
	if !domain.IsWithinRetention(createdAt, domain.RetentionFinancialDays, now) {
		t.Error("expected record from 5 years ago to be within 7-year retention")
	}
}

func TestIsWithinRetention_OutsidePeriod(t *testing.T) {
	// Record from 8 years ago with 7-year retention policy: outside retention
	now := time.Date(2026, 4, 9, 0, 0, 0, 0, time.UTC)
	createdAt := now.AddDate(-8, 0, 0)
	if domain.IsWithinRetention(createdAt, domain.RetentionFinancialDays, now) {
		t.Error("expected record from 8 years ago to be outside 7-year retention")
	}
}

func TestIsWithinRetention_ExactBoundary(t *testing.T) {
	// Record at exactly the retention boundary: createdAt + retentionDays == now
	// IsWithinRetention uses now.Before(retentionEnd), so at exact boundary it returns false
	now := time.Date(2026, 4, 9, 0, 0, 0, 0, time.UTC)
	createdAt := now.AddDate(0, 0, -domain.RetentionFinancialDays)
	// retentionEnd = createdAt + retentionDays = now, so now.Before(now) == false
	if domain.IsWithinRetention(createdAt, domain.RetentionFinancialDays, now) {
		t.Error("expected record at exact boundary to be outside retention (now is not before retentionEnd)")
	}
}

func TestRetentionConstants(t *testing.T) {
	if domain.RetentionFinancialYears != 7 {
		t.Errorf("expected RetentionFinancialYears == 7, got %d", domain.RetentionFinancialYears)
	}
	if domain.RetentionAccessLogYears != 2 {
		t.Errorf("expected RetentionAccessLogYears == 2, got %d", domain.RetentionAccessLogYears)
	}
	if domain.RetentionFinancialDays != 2555 {
		t.Errorf("expected RetentionFinancialDays == 2555, got %d", domain.RetentionFinancialDays)
	}
	if domain.RetentionAccessLogDays != 730 {
		t.Errorf("expected RetentionAccessLogDays == 730, got %d", domain.RetentionAccessLogDays)
	}
}
