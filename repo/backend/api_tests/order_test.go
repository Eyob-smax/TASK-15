package api_tests

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestOrders_MemberCanCreateListAndViewOnlyOwnOrders(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	memberA := app.seedUser(t, "member", nil)
	memberB := app.seedUser(t, "member", nil)
	item := app.seedItem(t, itemSeedOptions{
		CreatedBy: admin.ID,
		Name:      "Orderable Kettlebell",
		Status:    "published",
		Quantity:  20,
		AvailabilityWindows: []timeWindow{{
			Start: time.Now().UTC().Add(-1 * time.Hour),
			End:   time.Now().UTC().Add(48 * time.Hour),
		}},
	})

	memberACookies := app.login(t, memberA)
	memberBCookies := app.login(t, memberB)

	createARec := app.post(t, "/api/v1/orders", map[string]any{
		"item_id":  item.ID.String(),
		"quantity": 2,
	}, memberACookies)
	requireStatus(t, createARec, http.StatusCreated)
	orderA := decodeSuccess[map[string]any](t, createARec)

	createBRec := app.post(t, "/api/v1/orders", map[string]any{
		"item_id":  item.ID.String(),
		"quantity": 1,
	}, memberBCookies)
	requireStatus(t, createBRec, http.StatusCreated)
	orderB := decodeSuccess[map[string]any](t, createBRec)

	listRec := app.get(t, "/api/v1/orders", memberACookies)
	requireStatus(t, listRec, http.StatusOK)
	orders, _ := decodePaginated[map[string]any](t, listRec)
	if len(orders) != 1 || orders[0]["id"] != orderA["id"] {
		t.Fatalf("expected member to list only own order, got %#v", orders)
	}

	getOwnRec := app.get(t, "/api/v1/orders/"+orderA["id"].(string), memberACookies)
	requireStatus(t, getOwnRec, http.StatusOK)

	getOtherRec := app.get(t, "/api/v1/orders/"+orderB["id"].(string), memberACookies)
	requireStatus(t, getOtherRec, http.StatusForbidden)
}

func TestOrders_ManagerCanPayNoteRefundAndReadTimeline(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	manager := app.seedUser(t, "operations_manager", nil)
	member := app.seedUser(t, "member", nil)
	item := app.seedItem(t, itemSeedOptions{
		CreatedBy: admin.ID,
		Name:      "Refundable Bench",
		Status:    "published",
		Quantity:  10,
		AvailabilityWindows: []timeWindow{{
			Start: time.Now().UTC().Add(-1 * time.Hour),
			End:   time.Now().UTC().Add(48 * time.Hour),
		}},
	})

	orderRec := app.post(t, "/api/v1/orders", map[string]any{
		"item_id":  item.ID.String(),
		"quantity": 1,
	}, app.login(t, member))
	requireStatus(t, orderRec, http.StatusCreated)
	order := decodeSuccess[map[string]any](t, orderRec)
	orderID := order["id"].(string)

	managerCookies := app.login(t, manager)

	payRec := app.post(t, "/api/v1/orders/"+orderID+"/pay", map[string]any{
		"settlement_marker": "offline-cash-001",
	}, managerCookies)
	requireStatus(t, payRec, http.StatusOK)

	noteRec := app.post(t, "/api/v1/orders/"+orderID+"/notes", map[string]any{
		"note": "Split pickup requested",
	}, managerCookies)
	requireStatus(t, noteRec, http.StatusOK)

	refundRec := app.post(t, "/api/v1/orders/"+orderID+"/refund", map[string]any{}, managerCookies)
	requireStatus(t, refundRec, http.StatusOK)

	timelineRec := app.get(t, "/api/v1/orders/"+orderID+"/timeline", managerCookies)
	requireStatus(t, timelineRec, http.StatusOK)
	if strings.Contains(timelineRec.Body.String(), "Split pickup requested") {
		t.Fatal("expected timeline response to redact raw note content")
	}

	entries := decodeSuccess[[]map[string]any](t, timelineRec)
	if len(entries) < 3 {
		t.Fatalf("expected multiple timeline entries, got %#v", entries)
	}

	foundRedactedNote := false
	for _, entry := range entries {
		description, _ := entry["description"].(string)
		if description == "note added (content redacted)" {
			foundRedactedNote = true
			break
		}
	}
	if !foundRedactedNote {
		t.Fatalf("expected redacted note timeline entry, got %#v", entries)
	}
}

func TestOrders_ManagerCanSplitOrder(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	manager := app.seedUser(t, "operations_manager", nil)
	member := app.seedUser(t, "member", nil)
	item := app.seedItem(t, itemSeedOptions{
		CreatedBy: admin.ID,
		Name:      "Splittable Barbell",
		Status:    "published",
		Quantity:  10,
		AvailabilityWindows: []timeWindow{{
			Start: time.Now().UTC().Add(-1 * time.Hour),
			End:   time.Now().UTC().Add(48 * time.Hour),
		}},
	})

	orderRec := app.post(t, "/api/v1/orders", map[string]any{
		"item_id":  item.ID.String(),
		"quantity": 3,
	}, app.login(t, member))
	requireStatus(t, orderRec, http.StatusCreated)
	order := decodeSuccess[map[string]any](t, orderRec)
	orderID := order["id"].(string)

	managerCookies := app.login(t, manager)

	// Happy path: manager splits 3 into [1, 2].
	splitRec := app.post(t, "/api/v1/orders/"+orderID+"/split", map[string]any{
		"quantities": []int{1, 2},
	}, managerCookies)
	requireStatus(t, splitRec, http.StatusCreated)
	parts := decodeSuccess[[]map[string]any](t, splitRec)
	if len(parts) != 2 {
		t.Fatalf("expected 2 split orders, got %d", len(parts))
	}
	var totalQty int
	for _, p := range parts {
		totalQty += int(p["quantity"].(float64))
	}
	if totalQty != 3 {
		t.Fatalf("expected split quantities to sum to 3, got %d", totalQty)
	}

	// Member cannot split an order.
	anotherOrderRec := app.post(t, "/api/v1/orders", map[string]any{
		"item_id":  item.ID.String(),
		"quantity": 2,
	}, app.login(t, member))
	requireStatus(t, anotherOrderRec, http.StatusCreated)
	anotherOrder := decodeSuccess[map[string]any](t, anotherOrderRec)
	forbiddenRec := app.post(t, "/api/v1/orders/"+anotherOrder["id"].(string)+"/split", map[string]any{
		"quantities": []int{1, 1},
	}, app.login(t, member))
	requireStatus(t, forbiddenRec, http.StatusForbidden)

	// Invalid split: quantities don't sum to original quantity.
	badSplitRec := app.post(t, "/api/v1/orders/"+anotherOrder["id"].(string)+"/split", map[string]any{
		"quantities": []int{1, 5},
	}, managerCookies)
	if badSplitRec.Code == http.StatusCreated {
		t.Fatal("expected non-2xx for split with wrong quantity sum")
	}
}

func TestOrders_ManagerCanMergeOrders(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	manager := app.seedUser(t, "operations_manager", nil)
	member := app.seedUser(t, "member", nil)
	item := app.seedItem(t, itemSeedOptions{
		CreatedBy: admin.ID,
		Name:      "Mergeable Dumbbell",
		Status:    "published",
		Quantity:  10,
		AvailabilityWindows: []timeWindow{{
			Start: time.Now().UTC().Add(-1 * time.Hour),
			End:   time.Now().UTC().Add(48 * time.Hour),
		}},
	})

	memberCookies := app.login(t, member)

	order1Rec := app.post(t, "/api/v1/orders", map[string]any{
		"item_id":  item.ID.String(),
		"quantity": 2,
	}, memberCookies)
	requireStatus(t, order1Rec, http.StatusCreated)
	order1 := decodeSuccess[map[string]any](t, order1Rec)

	order2Rec := app.post(t, "/api/v1/orders", map[string]any{
		"item_id":  item.ID.String(),
		"quantity": 3,
	}, memberCookies)
	requireStatus(t, order2Rec, http.StatusCreated)
	order2 := decodeSuccess[map[string]any](t, order2Rec)

	managerCookies := app.login(t, manager)

	// Happy path: manager merges the two orders.
	mergeRec := app.post(t, "/api/v1/orders/merge", map[string]any{
		"order_ids": []string{order1["id"].(string), order2["id"].(string)},
	}, managerCookies)
	requireStatus(t, mergeRec, http.StatusCreated)
	merged := decodeSuccess[map[string]any](t, mergeRec)
	mergedQty := int(merged["quantity"].(float64))
	if mergedQty != 5 {
		t.Fatalf("expected merged order quantity 5, got %d", mergedQty)
	}

	// Member cannot merge orders.
	order3Rec := app.post(t, "/api/v1/orders", map[string]any{
		"item_id":  item.ID.String(),
		"quantity": 1,
	}, memberCookies)
	requireStatus(t, order3Rec, http.StatusCreated)
	order3 := decodeSuccess[map[string]any](t, order3Rec)

	order4Rec := app.post(t, "/api/v1/orders", map[string]any{
		"item_id":  item.ID.String(),
		"quantity": 1,
	}, memberCookies)
	requireStatus(t, order4Rec, http.StatusCreated)
	order4 := decodeSuccess[map[string]any](t, order4Rec)

	forbiddenRec := app.post(t, "/api/v1/orders/merge", map[string]any{
		"order_ids": []string{order3["id"].(string), order4["id"].(string)},
	}, memberCookies)
	requireStatus(t, forbiddenRec, http.StatusForbidden)
}
