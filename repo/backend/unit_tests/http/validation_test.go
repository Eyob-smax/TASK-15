package http_test

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"fitcommerce/internal/http/dto"
)

func TestLoginRequest_ValidJSON(t *testing.T) {
	raw := `{"email":"admin@fitcommerce.io","password":"securePass123!"}`
	var req dto.LoginRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if req.Email != "admin@fitcommerce.io" {
		t.Errorf("expected email 'admin@fitcommerce.io', got %q", req.Email)
	}
	if req.Password != "securePass123!" {
		t.Errorf("expected password 'securePass123!', got %q", req.Password)
	}
}

func TestCreateItemRequest_ValidJSON(t *testing.T) {
	raw := `{
		"name": "Power Rack",
		"description": "Heavy-duty steel power rack",
		"category": "strength",
		"brand": "Rogue",
		"condition": "new",
		"refundable_deposit": 100.00,
		"billing_model": "one_time",
		"quantity": 5,
		"availability_windows": [
			{"start_time": "2026-05-01T08:00:00Z", "end_time": "2026-05-31T18:00:00Z"}
		]
	}`
	var req dto.CreateItemRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if req.Name != "Power Rack" {
		t.Errorf("expected name 'Power Rack', got %q", req.Name)
	}
	if req.Quantity != 5 {
		t.Errorf("expected quantity 5, got %d", req.Quantity)
	}
	if len(req.AvailabilityWindows) != 1 {
		t.Fatalf("expected 1 availability window, got %d", len(req.AvailabilityWindows))
	}
}

func TestBatchEditRequest_MultipleRows(t *testing.T) {
	raw := `{
		"edits": [
			{"item_id": "abc-123", "field": "price", "new_value": "99.99"},
			{"item_id": "def-456", "field": "quantity", "new_value": "20"},
			{"item_id": "ghi-789", "field": "status", "new_value": "published"}
		]
	}`
	var req dto.BatchEditRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(req.Edits) != 3 {
		t.Errorf("expected 3 edits, got %d", len(req.Edits))
	}
}

func TestPaginationMeta_Calculation(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		pageSize   int
		totalCount int
		wantPages  int
	}{
		{"exact division", 1, 10, 100, 10},
		{"with remainder", 1, 10, 105, 11},
		{"single page", 1, 25, 5, 1},
		{"empty result", 1, 10, 0, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			totalPages := 0
			if tc.totalCount > 0 {
				totalPages = int(math.Ceil(float64(tc.totalCount) / float64(tc.pageSize)))
			}
			meta := dto.PaginationMeta{
				Page:       tc.page,
				PageSize:   tc.pageSize,
				TotalCount: tc.totalCount,
				TotalPages: totalPages,
			}
			if meta.TotalPages != tc.wantPages {
				t.Errorf("expected %d total pages, got %d", tc.wantPages, meta.TotalPages)
			}
		})
	}
}

func TestAvailabilityWindowRequest_TimeParsing(t *testing.T) {
	raw := `{"start_time":"2026-05-01T08:00:00Z","end_time":"2026-05-31T18:00:00Z"}`
	var req dto.AvailabilityWindowRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	expectedStart := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2026, 5, 31, 18, 0, 0, 0, time.UTC)
	if !req.StartTime.Equal(expectedStart) {
		t.Errorf("expected start_time %v, got %v", expectedStart, req.StartTime)
	}
	if !req.EndTime.Equal(expectedEnd) {
		t.Errorf("expected end_time %v, got %v", expectedEnd, req.EndTime)
	}
}
