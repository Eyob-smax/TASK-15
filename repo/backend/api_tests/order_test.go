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
