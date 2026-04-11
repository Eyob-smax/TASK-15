package api_tests

import (
	"net/http"
	"testing"
	"time"
)

func TestCampaigns_MemberCanStartCampaign(t *testing.T) {
	app := newIntegrationApp(t)
	member := app.seedUser(t, "member", nil)
	admin := app.seedUser(t, "administrator", nil)
	item := app.seedItem(t, itemSeedOptions{
		CreatedBy: admin.ID,
		Status:    "published",
		Quantity:  8,
		AvailabilityWindows: []timeWindow{{
			Start: time.Now().UTC().Add(-1 * time.Hour),
			End:   time.Now().UTC().Add(48 * time.Hour),
		}},
	})

	rec := app.post(t, "/api/v1/campaigns", map[string]any{
		"item_id":      item.ID.String(),
		"min_quantity": 3,
		"cutoff_time":  time.Now().UTC().Add(12 * time.Hour).Format(time.RFC3339),
	}, app.login(t, member))
	requireStatus(t, rec, http.StatusCreated)

	campaign := decodeSuccess[map[string]any](t, rec)
	if campaign["item_id"] != item.ID.String() {
		t.Fatalf("expected campaign item_id %s, got %#v", item.ID.String(), campaign["item_id"])
	}
	if campaign["created_by"] != member.ID.String() {
		t.Fatalf("expected member-created campaign, got %#v", campaign["created_by"])
	}
}

func TestCampaigns_OperationsManagerCanCreateAndMemberCanJoin(t *testing.T) {
	app := newIntegrationApp(t)
	ops := app.seedUser(t, "operations_manager", nil)
	member := app.seedUser(t, "member", nil)
	item := app.seedItem(t, itemSeedOptions{
		CreatedBy: ops.ID,
		Name:      "Campaign Barbell",
		Status:    "published",
		Quantity:  10,
		AvailabilityWindows: []timeWindow{{
			Start: time.Now().UTC().Add(-1 * time.Hour),
			End:   time.Now().UTC().Add(48 * time.Hour),
		}},
	})

	createRec := app.post(t, "/api/v1/campaigns", map[string]any{
		"item_id":      item.ID.String(),
		"min_quantity": 4,
		"cutoff_time":  time.Now().UTC().Add(12 * time.Hour).Format(time.RFC3339),
	}, app.login(t, ops))
	requireStatus(t, createRec, http.StatusCreated)

	campaign := decodeSuccess[map[string]any](t, createRec)
	campaignID := campaign["id"].(string)

	joinRec := app.post(t, "/api/v1/campaigns/"+campaignID+"/join", map[string]any{
		"quantity": 2,
	}, app.login(t, member))
	requireStatus(t, joinRec, http.StatusCreated)

	getRec := app.get(t, "/api/v1/campaigns/"+campaignID, app.login(t, member))
	requireStatus(t, getRec, http.StatusOK)

	got := decodeSuccess[map[string]any](t, getRec)
	if int(got["current_committed_qty"].(float64)) != 2 {
		t.Fatalf("expected current_committed_qty 2, got %#v", got["current_committed_qty"])
	}
}
