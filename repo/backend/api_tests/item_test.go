package api_tests

import (
	"net/http"
	"testing"
	"time"
)

func TestItems_CreateGetAndUpdateRespectVersioning(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	adminCookies := app.login(t, admin)

	start := time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)
	end := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	createRec := app.post(t, "/api/v1/items", map[string]any{
		"sku":          "SKU-RACK-01",
		"name":         "Power Rack",
		"description":  "Heavy-duty power rack",
		"category":     "strength",
		"brand":        "Rogue",
		"condition":    "new",
		"unit_price":   899.99,
		"billing_model": "one_time",
		"quantity":     8,
		"availability_windows": []map[string]string{
			{"start_time": start, "end_time": end},
		},
	}, adminCookies)
	requireStatus(t, createRec, http.StatusCreated)

	created := decodeSuccess[map[string]any](t, createRec)
	itemID := created["id"].(string)

	getRec := app.get(t, "/api/v1/items/"+itemID, adminCookies)
	requireStatus(t, getRec, http.StatusOK)

	item := decodeSuccess[map[string]any](t, getRec)
	if item["name"] != "Power Rack" {
		t.Fatalf("expected created item name, got %#v", item["name"])
	}
	if len(item["availability_windows"].([]any)) != 1 {
		t.Fatalf("expected availability windows on item detail, got %#v", item["availability_windows"])
	}

	updateRec := app.put(t, "/api/v1/items/"+itemID, map[string]any{
		"name":     "Power Rack Elite",
		"quantity": 10,
		"version":  1,
	}, adminCookies)
	requireStatus(t, updateRec, http.StatusOK)

	updated := decodeSuccess[map[string]any](t, updateRec)
	if updated["name"] != "Power Rack Elite" {
		t.Fatalf("expected updated name, got %#v", updated["name"])
	}
	if int(updated["version"].(float64)) != 2 {
		t.Fatalf("expected version 2 after update, got %#v", updated["version"])
	}

	staleRec := app.put(t, "/api/v1/items/"+itemID, map[string]any{
		"name":    "Stale Update",
		"version": 1,
	}, adminCookies)
	requireStatus(t, staleRec, http.StatusConflict)
}

func TestItems_PublishAndBatchEditUseRealValidation(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	member := app.seedUser(t, "member", nil)
	adminCookies := app.login(t, admin)

	memberCreate := app.post(t, "/api/v1/items", map[string]any{
		"name":          "Forbidden Item",
		"category":      "strength",
		"brand":         "FitCommerce",
		"condition":     "new",
		"billing_model": "one_time",
		"quantity":      1,
	}, app.login(t, member))
	requireStatus(t, memberCreate, http.StatusForbidden)

	overlapStart := time.Now().UTC().Add(-30 * time.Minute).Format(time.RFC3339)
	overlapEnd := time.Now().UTC().Add(4 * time.Hour).Format(time.RFC3339)

	createBlockedRec := app.post(t, "/api/v1/items", map[string]any{
		"name":          "Blocked Publish Item",
		"category":      "supplements",
		"brand":         "Core Labs",
		"condition":     "new",
		"billing_model": "one_time",
		"quantity":      5,
		"availability_windows": []map[string]string{
			{"start_time": overlapStart, "end_time": overlapEnd},
		},
		"blackout_windows": []map[string]string{
			{"start_time": overlapStart, "end_time": overlapEnd},
		},
	}, adminCookies)
	requireStatus(t, createBlockedRec, http.StatusCreated)

	blockedItem := decodeSuccess[map[string]any](t, createBlockedRec)
	blockedPublishRec := app.post(t, "/api/v1/items/"+blockedItem["id"].(string)+"/publish", map[string]any{}, adminCookies)
	requireStatus(t, blockedPublishRec, http.StatusUnprocessableEntity)
	if decodeError(t, blockedPublishRec).Error.Code != "PUBLISH_BLOCKED" {
		t.Fatal("expected PUBLISH_BLOCKED when publish windows overlap")
	}

	validStart := time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)
	validEnd := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	createValidRec := app.post(t, "/api/v1/items", map[string]any{
		"name":          "Batch Edit Item",
		"category":      "strength",
		"brand":         "FitCommerce",
		"condition":     "new",
		"billing_model": "one_time",
		"quantity":      4,
		"availability_windows": []map[string]string{
			{"start_time": validStart, "end_time": validEnd},
		},
	}, adminCookies)
	requireStatus(t, createValidRec, http.StatusCreated)

	validItem := decodeSuccess[map[string]any](t, createValidRec)
	validID := validItem["id"].(string)

	publishRec := app.post(t, "/api/v1/items/"+validID+"/publish", map[string]any{}, adminCookies)
	requireStatus(t, publishRec, http.StatusOK)

	listRec := app.get(t, "/api/v1/items?status=published", adminCookies)
	requireStatus(t, listRec, http.StatusOK)
	listed, _ := decodePaginated[map[string]any](t, listRec)
	if len(listed) == 0 {
		t.Fatal("expected list endpoint to return published items")
	}

	unpublishRec := app.post(t, "/api/v1/items/"+validID+"/unpublish", map[string]any{}, adminCookies)
	requireStatus(t, unpublishRec, http.StatusOK)

	batchRec := app.post(t, "/api/v1/items/batch-edit", map[string]any{
		"edits": []map[string]string{
			{"item_id": validID, "field": "unit_price", "new_value": "129.99"},
			{"item_id": validID, "field": "quantity", "new_value": "-5"},
		},
	}, adminCookies)
	requireStatus(t, batchRec, http.StatusMultiStatus)

	batch := decodeSuccess[map[string]any](t, batchRec)
	if int(batch["success_count"].(float64)) != 1 || int(batch["failure_count"].(float64)) != 1 {
		t.Fatalf("expected one success and one failure, got %#v", batch)
	}

	availabilityBatchRec := app.post(t, "/api/v1/items/batch-edit", map[string]any{
		"edits": []map[string]any{
			{
				"item_id": validID,
				"field":   "availability_windows",
				"availability_windows": []map[string]string{
					{
						"start_time": time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339),
						"end_time":   time.Now().UTC().Add(6 * time.Hour).Format(time.RFC3339),
					},
				},
			},
			{
				"item_id": validID,
				"field":   "availability_windows",
				"availability_windows": []map[string]string{
					{
						"start_time": time.Now().UTC().Add(8 * time.Hour).Format(time.RFC3339),
						"end_time":   time.Now().UTC().Add(7 * time.Hour).Format(time.RFC3339),
					},
				},
			},
		},
	}, adminCookies)
	requireStatus(t, availabilityBatchRec, http.StatusMultiStatus)

	availabilityBatch := decodeSuccess[map[string]any](t, availabilityBatchRec)
	if int(availabilityBatch["success_count"].(float64)) != 1 || int(availabilityBatch["failure_count"].(float64)) != 1 {
		t.Fatalf("expected availability batch to record one success and one failure, got %#v", availabilityBatch)
	}
}
