package api_tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHTTP_ItemCampaignAuthAndLocationBranches(t *testing.T) {
	app := newIntegrationApp(t)

	locationA := app.seedLocation(t, "Branch A")
	locationB := app.seedLocation(t, "Branch B")
	admin := app.seedUser(t, "administrator", nil)
	ops := app.seedUser(t, "operations_manager", &locationA.ID)
	member := app.seedUser(t, "member", nil)

	adminCookies := app.login(t, admin)
	opsCookies := app.login(t, ops)
	memberCookies := app.login(t, member)

	// Cover GetSession response branch that includes user.location_id.
	sessionRec := app.get(t, "/api/v1/auth/session", opsCookies)
	requireStatus(t, sessionRec, http.StatusOK)
	session := decodeSuccess[map[string]any](t, sessionRec)
	sessionUser := session["user"].(map[string]any)
	if sessionUser["location_id"] != locationA.ID.String() {
		t.Fatalf("expected session user location_id %s, got %#v", locationA.ID.String(), sessionUser["location_id"])
	}

	// Cover non-admin same-location allow branch.
	ownLocationRec := app.get(t, "/api/v1/locations/"+locationA.ID.String(), opsCookies)
	requireStatus(t, ownLocationRec, http.StatusOK)

	badCreateBodyRec := app.post(t, "/api/v1/items", "not-an-object", adminCookies)
	requireStatus(t, badCreateBodyRec, http.StatusBadRequest)

	badCreateLocationRec := app.post(t, "/api/v1/items", map[string]any{
		"name":           "Bad Create Location",
		"category":       "strength",
		"brand":          "FitCommerce",
		"condition":      "new",
		"billing_model":  "one_time",
		"quantity":       1,
		"location_id":    "not-a-uuid",
		"availability_windows": []map[string]string{{
			"start_time": time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339),
			"end_time":   time.Now().UTC().Add(3 * time.Hour).Format(time.RFC3339),
		}},
	}, adminCookies)
	requireStatus(t, badCreateLocationRec, http.StatusBadRequest)

	createRec := app.post(t, "/api/v1/items", map[string]any{
		"sku":               "SKU-BRANCH-COVER-01",
		"name":              "Branch Coverage Item",
		"description":       "Initial",
		"category":          "strength",
		"brand":             "FitCommerce",
		"condition":         "new",
		"unit_price":        100.0,
		"refundable_deposit": 10.0,
		"billing_model":     "one_time",
		"quantity":          6,
		"location_id":       locationA.ID.String(),
		"availability_windows": []map[string]string{{
			"start_time": time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339),
			"end_time":   time.Now().UTC().Add(4 * time.Hour).Format(time.RFC3339),
		}},
		"blackout_windows": []map[string]string{{
			"start_time": time.Now().UTC().Add(5 * time.Hour).Format(time.RFC3339),
			"end_time":   time.Now().UTC().Add(6 * time.Hour).Format(time.RFC3339),
		}},
	}, adminCookies)
	requireStatus(t, createRec, http.StatusCreated)
	itemID := decodeSuccess[map[string]any](t, createRec)["id"].(string)

	updateRec := app.put(t, "/api/v1/items/"+itemID, map[string]any{
		"sku":               "SKU-BRANCH-COVER-02",
		"name":              "Branch Coverage Item Updated",
		"description":       "Updated",
		"category":          "conditioning",
		"brand":             "UpdatedBrand",
		"condition":         "used",
		"unit_price":        111.0,
		"refundable_deposit": 12.0,
		"billing_model":     "monthly_rental",
		"quantity":          9,
		"location_id":       locationB.ID.String(),
		"version":           1,
		"availability_windows": []map[string]string{{
			"start_time": time.Now().UTC().Add(-30 * time.Minute).Format(time.RFC3339),
			"end_time":   time.Now().UTC().Add(8 * time.Hour).Format(time.RFC3339),
		}},
		"blackout_windows": []map[string]string{{
			"start_time": time.Now().UTC().Add(9 * time.Hour).Format(time.RFC3339),
			"end_time":   time.Now().UTC().Add(10 * time.Hour).Format(time.RFC3339),
		}},
	}, adminCookies)
	requireStatus(t, updateRec, http.StatusOK)

	badUpdateBodyRec := app.put(t, "/api/v1/items/"+itemID, "not-an-object", adminCookies)
	requireStatus(t, badUpdateBodyRec, http.StatusBadRequest)

	badUpdateLocationRec := app.put(t, "/api/v1/items/"+itemID, map[string]any{
		"location_id": "not-a-uuid",
		"version":     2,
	}, adminCookies)
	requireStatus(t, badUpdateLocationRec, http.StatusBadRequest)

	badGetItemRec := app.get(t, "/api/v1/items/not-a-uuid", adminCookies)
	requireStatus(t, badGetItemRec, http.StatusBadRequest)

	badCampaignCreateBodyRec := app.post(t, "/api/v1/campaigns", "not-an-object", adminCookies)
	requireStatus(t, badCampaignCreateBodyRec, http.StatusBadRequest)

	badCampaignItemIDRec := app.post(t, "/api/v1/campaigns", map[string]any{
		"item_id":      "not-a-uuid",
		"min_quantity": 2,
		"cutoff_time":  time.Now().UTC().Add(4 * time.Hour).Format(time.RFC3339),
	}, adminCookies)
	requireStatus(t, badCampaignItemIDRec, http.StatusBadRequest)

	campaignItem := app.seedItem(t, itemSeedOptions{
		CreatedBy: admin.ID,
		Name:      "Campaign Branch Coverage",
		Status:    "published",
		Quantity:  15,
		AvailabilityWindows: []timeWindow{{
			Start: time.Now().UTC().Add(-1 * time.Hour),
			End:   time.Now().UTC().Add(24 * time.Hour),
		}},
	})

	createCampaignRec := app.post(t, "/api/v1/campaigns", map[string]any{
		"item_id":      campaignItem.ID.String(),
		"min_quantity": 3,
		"cutoff_time":  time.Now().UTC().Add(6 * time.Hour).Format(time.RFC3339),
	}, adminCookies)
	requireStatus(t, createCampaignRec, http.StatusCreated)
	campaignID := decodeSuccess[map[string]any](t, createCampaignRec)["id"].(string)

	badCampaignGetRec := app.get(t, "/api/v1/campaigns/not-a-uuid", adminCookies)
	requireStatus(t, badCampaignGetRec, http.StatusBadRequest)

	badJoinIDRec := app.post(t, "/api/v1/campaigns/not-a-uuid/join", map[string]any{"quantity": 1}, memberCookies)
	requireStatus(t, badJoinIDRec, http.StatusBadRequest)

	badJoinBodyRec := app.post(t, "/api/v1/campaigns/"+campaignID+"/join", "not-an-object", memberCookies)
	requireStatus(t, badJoinBodyRec, http.StatusBadRequest)

	badJoinValidationRec := app.post(t, "/api/v1/campaigns/"+campaignID+"/join", map[string]any{}, memberCookies)
	requireStatus(t, badJoinValidationRec, http.StatusUnprocessableEntity)
}

func TestHTTP_ProcurementVarianceInventoryReportAndRetentionBranches(t *testing.T) {
	app := newIntegrationApp(t)

	admin := app.seedUser(t, "administrator", nil)
	procurement := app.seedUser(t, "procurement_specialist", nil)
	adminCookies := app.login(t, admin)
	procurementCookies := app.login(t, procurement)

	badSupplierBodyRec := app.post(t, "/api/v1/suppliers", "not-an-object", procurementCookies)
	requireStatus(t, badSupplierBodyRec, http.StatusBadRequest)

	badSupplierValidationRec := app.post(t, "/api/v1/suppliers", map[string]any{}, procurementCookies)
	requireStatus(t, badSupplierValidationRec, http.StatusUnprocessableEntity)

	supplierCreateRec := app.post(t, "/api/v1/suppliers", map[string]any{
		"name": "Coverage Supplier",
	}, procurementCookies)
	requireStatus(t, supplierCreateRec, http.StatusCreated)
	supplierID := decodeSuccess[map[string]any](t, supplierCreateRec)["id"].(string)

	badSupplierUpdateBodyRec := app.put(t, "/api/v1/suppliers/"+supplierID, "not-an-object", procurementCookies)
	requireStatus(t, badSupplierUpdateBodyRec, http.StatusBadRequest)

	badCreatePOSupplierRec := app.post(t, "/api/v1/purchase-orders", map[string]any{
		"supplier_id": "not-a-uuid",
		"lines": []map[string]any{{
			"item_id":            uuid.New().String(),
			"ordered_quantity":   1,
			"ordered_unit_price": 3.5,
		}},
	}, procurementCookies)
	requireStatus(t, badCreatePOSupplierRec, http.StatusBadRequest)

	badCreatePOItemRec := app.post(t, "/api/v1/purchase-orders", map[string]any{
		"supplier_id": uuid.New().String(),
		"lines": []map[string]any{{
			"item_id":            "not-a-uuid",
			"ordered_quantity":   1,
			"ordered_unit_price": 3.5,
		}},
	}, procurementCookies)
	requireStatus(t, badCreatePOItemRec, http.StatusBadRequest)

	badApproveIDRec := app.post(t, "/api/v1/purchase-orders/not-a-uuid/approve", map[string]any{}, procurementCookies)
	requireStatus(t, badApproveIDRec, http.StatusBadRequest)

	badReceiveIDRec := app.post(t, "/api/v1/purchase-orders/not-a-uuid/receive", map[string]any{
		"lines": []map[string]any{{
			"po_line_id":          uuid.New().String(),
			"received_quantity":   1,
			"received_unit_price": 1.0,
		}},
	}, procurementCookies)
	requireStatus(t, badReceiveIDRec, http.StatusBadRequest)

	badReceiveBodyRec := app.post(t, "/api/v1/purchase-orders/"+uuid.New().String()+"/receive", "not-an-object", procurementCookies)
	requireStatus(t, badReceiveBodyRec, http.StatusBadRequest)

	badReceiveLineIDRec := app.post(t, "/api/v1/purchase-orders/"+uuid.New().String()+"/receive", map[string]any{
		"lines": []map[string]any{{
			"po_line_id":          "not-a-uuid",
			"received_quantity":   1,
			"received_unit_price": 1.0,
		}},
	}, procurementCookies)
	requireStatus(t, badReceiveLineIDRec, http.StatusBadRequest)

	badVarianceGetRec := app.get(t, "/api/v1/variances/not-a-uuid", procurementCookies)
	requireStatus(t, badVarianceGetRec, http.StatusBadRequest)

	badVarianceResolveIDRec := app.post(t, "/api/v1/variances/not-a-uuid/resolve", map[string]any{
		"action":           "return",
		"resolution_notes": "invalid id branch",
	}, procurementCookies)
	requireStatus(t, badVarianceResolveIDRec, http.StatusBadRequest)

	badVarianceResolveBodyRec := app.post(t, "/api/v1/variances/"+uuid.New().String()+"/resolve", "not-an-object", procurementCookies)
	requireStatus(t, badVarianceResolveBodyRec, http.StatusBadRequest)

	adjustmentWithoutQuantityRec := app.post(t, "/api/v1/variances/"+uuid.New().String()+"/resolve", map[string]any{
		"action":           "adjustment",
		"resolution_notes": "missing quantity",
	}, procurementCookies)
	requireStatus(t, adjustmentWithoutQuantityRec, http.StatusUnprocessableEntity)

	returnWithQuantityRec := app.post(t, "/api/v1/variances/"+uuid.New().String()+"/resolve", map[string]any{
		"action":           "return",
		"resolution_notes": "return cannot include quantity",
		"quantity_change":  1,
	}, procurementCookies)
	requireStatus(t, returnWithQuantityRec, http.StatusUnprocessableEntity)

	badSnapshotsLocationRec := app.get(t, "/api/v1/inventory/snapshots?location_id=not-a-uuid", procurementCookies)
	requireStatus(t, badSnapshotsLocationRec, http.StatusBadRequest)

	badAdjustmentBodyRec := app.post(t, "/api/v1/inventory/adjustments", "not-an-object", adminCookies)
	requireStatus(t, badAdjustmentBodyRec, http.StatusBadRequest)

	badAdjustmentItemRec := app.post(t, "/api/v1/inventory/adjustments", map[string]any{
		"item_id":         "not-a-uuid",
		"quantity_change": 1,
		"reason":          "branch coverage",
	}, adminCookies)
	requireStatus(t, badAdjustmentItemRec, http.StatusBadRequest)

	badRunExportBodyRec := app.post(t, "/api/v1/exports", "not-an-object", adminCookies)
	requireStatus(t, badRunExportBodyRec, http.StatusBadRequest)

	badRunExportReportIDRec := app.post(t, "/api/v1/exports", map[string]any{
		"report_id": "not-a-uuid",
		"format":    "csv",
	}, adminCookies)
	requireStatus(t, badRunExportReportIDRec, http.StatusBadRequest)

	badExportGetIDRec := app.get(t, "/api/v1/exports/not-a-uuid", adminCookies)
	requireStatus(t, badExportGetIDRec, http.StatusBadRequest)

	badExportDownloadIDRec := app.get(t, "/api/v1/exports/not-a-uuid/download", adminCookies)
	requireStatus(t, badExportDownloadIDRec, http.StatusBadRequest)

	badRetentionBodyRec := app.put(t, "/api/v1/admin/retention-policies/access_logs", "not-an-object", adminCookies)
	requireStatus(t, badRetentionBodyRec, http.StatusBadRequest)

	badRetentionValidationRec := app.put(t, "/api/v1/admin/retention-policies/access_logs", map[string]any{}, adminCookies)
	requireStatus(t, badRetentionValidationRec, http.StatusUnprocessableEntity)
}