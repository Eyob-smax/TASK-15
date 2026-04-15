package domain_test

import (
	"testing"

	"fitcommerce/internal/domain"
)

// ─── UserStatus ────────────────────────────────────────────────────────────────

func TestUserStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status domain.UserStatus
		valid  bool
	}{
		{"active is valid", domain.UserStatusActive, true},
		{"inactive is valid", domain.UserStatusInactive, true},
		{"locked is valid", domain.UserStatusLocked, true},
		{"deleted is invalid", domain.UserStatus("deleted"), false},
		{"empty is invalid", domain.UserStatus(""), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.status.IsValid(); got != tc.valid {
				t.Errorf("UserStatus(%q).IsValid() = %v, want %v", tc.status, got, tc.valid)
			}
		})
	}
}

func TestAllUserStatuses_ReturnsAll(t *testing.T) {
	statuses := domain.AllUserStatuses()
	if len(statuses) != 3 {
		t.Errorf("expected 3 user statuses, got %d", len(statuses))
	}
}

// ─── VarianceType ──────────────────────────────────────────────────────────────

func TestVarianceType_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		vtype domain.VarianceType
		valid bool
	}{
		{"shortage is valid", domain.VarianceTypeShortage, true},
		{"overage is valid", domain.VarianceTypeOverage, true},
		{"price_difference is valid", domain.VarianceTypePriceDifference, true},
		{"unknown is invalid", domain.VarianceType("unknown"), false},
		{"empty is invalid", domain.VarianceType(""), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.vtype.IsValid(); got != tc.valid {
				t.Errorf("VarianceType(%q).IsValid() = %v, want %v", tc.vtype, got, tc.valid)
			}
		})
	}
}

func TestAllVarianceTypes_ReturnsAll(t *testing.T) {
	types := domain.AllVarianceTypes()
	if len(types) != 3 {
		t.Errorf("expected 3 variance types, got %d", len(types))
	}
}

// ─── VarianceStatus ────────────────────────────────────────────────────────────

func TestVarianceStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status domain.VarianceStatus
		valid  bool
	}{
		{"open is valid", domain.VarianceStatusOpen, true},
		{"resolved is valid", domain.VarianceStatusResolved, true},
		{"escalated is valid", domain.VarianceStatusEscalated, true},
		{"closed is invalid", domain.VarianceStatus("closed"), false},
		{"empty is invalid", domain.VarianceStatus(""), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.status.IsValid(); got != tc.valid {
				t.Errorf("VarianceStatus(%q).IsValid() = %v, want %v", tc.status, got, tc.valid)
			}
		})
	}
}

func TestAllVarianceStatuses_ReturnsAll(t *testing.T) {
	statuses := domain.AllVarianceStatuses()
	if len(statuses) != 3 {
		t.Errorf("expected 3 variance statuses, got %d", len(statuses))
	}
}

// ─── MembershipStatus ──────────────────────────────────────────────────────────

func TestMembershipStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status domain.MembershipStatus
		valid  bool
	}{
		{"active is valid", domain.MembershipStatusActive, true},
		{"expired is valid", domain.MembershipStatusExpired, true},
		{"cancelled is valid", domain.MembershipStatusCancelled, true},
		{"suspended is valid", domain.MembershipStatusSuspended, true},
		{"frozen is invalid", domain.MembershipStatus("frozen"), false},
		{"empty is invalid", domain.MembershipStatus(""), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.status.IsValid(); got != tc.valid {
				t.Errorf("MembershipStatus(%q).IsValid() = %v, want %v", tc.status, got, tc.valid)
			}
		})
	}
}

func TestAllMembershipStatuses_ReturnsAll(t *testing.T) {
	statuses := domain.AllMembershipStatuses()
	if len(statuses) != 4 {
		t.Errorf("expected 4 membership statuses, got %d", len(statuses))
	}
}

// ─── ExportFormat ──────────────────────────────────────────────────────────────

func TestAllExportFormats_ReturnsAll(t *testing.T) {
	formats := domain.AllExportFormats()
	if len(formats) != 2 {
		t.Errorf("expected 2 export formats, got %d", len(formats))
	}
}

// ─── ExportStatus ──────────────────────────────────────────────────────────────

func TestExportStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status domain.ExportStatus
		valid  bool
	}{
		{"pending is valid", domain.ExportStatusPending, true},
		{"processing is valid", domain.ExportStatusProcessing, true},
		{"completed is valid", domain.ExportStatusCompleted, true},
		{"failed is valid", domain.ExportStatusFailed, true},
		{"done is invalid", domain.ExportStatus("done"), false},
		{"empty is invalid", domain.ExportStatus(""), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.status.IsValid(); got != tc.valid {
				t.Errorf("ExportStatus(%q).IsValid() = %v, want %v", tc.status, got, tc.valid)
			}
		})
	}
}

func TestAllExportStatuses_ReturnsAll(t *testing.T) {
	statuses := domain.AllExportStatuses()
	if len(statuses) != 4 {
		t.Errorf("expected 4 export statuses, got %d", len(statuses))
	}
}

// ─── BackupStatus ──────────────────────────────────────────────────────────────

func TestBackupStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status domain.BackupStatus
		valid  bool
	}{
		{"running is valid", domain.BackupStatusRunning, true},
		{"completed is valid", domain.BackupStatusCompleted, true},
		{"failed is valid", domain.BackupStatusFailed, true},
		{"pending is invalid", domain.BackupStatus("pending"), false},
		{"empty is invalid", domain.BackupStatus(""), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.status.IsValid(); got != tc.valid {
				t.Errorf("BackupStatus(%q).IsValid() = %v, want %v", tc.status, got, tc.valid)
			}
		})
	}
}

func TestAllBackupStatuses_ReturnsAll(t *testing.T) {
	statuses := domain.AllBackupStatuses()
	if len(statuses) != 3 {
		t.Errorf("expected 3 backup statuses, got %d", len(statuses))
	}
}

// ─── EncryptionKeyStatus ───────────────────────────────────────────────────────

func TestEncryptionKeyStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status domain.EncryptionKeyStatus
		valid  bool
	}{
		{"active is valid", domain.EncryptionKeyStatusActive, true},
		{"rotated is valid", domain.EncryptionKeyStatusRotated, true},
		{"revoked is valid", domain.EncryptionKeyStatusRevoked, true},
		{"expired is invalid", domain.EncryptionKeyStatus("expired"), false},
		{"empty is invalid", domain.EncryptionKeyStatus(""), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.status.IsValid(); got != tc.valid {
				t.Errorf("EncryptionKeyStatus(%q).IsValid() = %v, want %v", tc.status, got, tc.valid)
			}
		})
	}
}

func TestAllEncryptionKeyStatuses_ReturnsAll(t *testing.T) {
	statuses := domain.AllEncryptionKeyStatuses()
	if len(statuses) != 3 {
		t.Errorf("expected 3 encryption key statuses, got %d", len(statuses))
	}
}
