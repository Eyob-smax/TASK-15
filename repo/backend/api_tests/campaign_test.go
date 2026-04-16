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

func TestCampaigns_ListCancelAndEvaluateEndpoints(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	member := app.seedUser(t, "member", nil)
	adminCookies := app.login(t, admin)

	itemA := app.seedItem(t, itemSeedOptions{
		CreatedBy: admin.ID,
		Name:      "Eval Target",
		Status:    "published",
		Quantity:  9,
		AvailabilityWindows: []timeWindow{{
			Start: time.Now().UTC().Add(-2 * time.Hour),
			End:   time.Now().UTC().Add(24 * time.Hour),
		}},
	})

	evalCreateRec := app.post(t, "/api/v1/campaigns", map[string]any{
		"item_id":      itemA.ID.String(),
		"min_quantity": 5,
		"cutoff_time":  time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339),
	}, adminCookies)
	requireStatus(t, evalCreateRec, http.StatusCreated)
	evalCampaignID := decodeSuccess[map[string]any](t, evalCreateRec)["id"].(string)

	itemB := app.seedItem(t, itemSeedOptions{
		CreatedBy: admin.ID,
		Name:      "Cancel Target",
		Status:    "published",
		Quantity:  7,
		AvailabilityWindows: []timeWindow{{
			Start: time.Now().UTC().Add(-2 * time.Hour),
			End:   time.Now().UTC().Add(24 * time.Hour),
		}},
	})

	cancelCreateRec := app.post(t, "/api/v1/campaigns", map[string]any{
		"item_id":      itemB.ID.String(),
		"min_quantity": 2,
		"cutoff_time":  time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339),
	}, adminCookies)
	requireStatus(t, cancelCreateRec, http.StatusCreated)
	cancelCampaignID := decodeSuccess[map[string]any](t, cancelCreateRec)["id"].(string)

	listRec := app.get(t, "/api/v1/campaigns", adminCookies)
	requireStatus(t, listRec, http.StatusOK)
	listRows, _ := decodePaginated[map[string]any](t, listRec)
	if len(listRows) == 0 {
		t.Fatal("expected campaigns list to include created records")
	}

	evaluateRec := app.post(t, "/api/v1/campaigns/"+evalCampaignID+"/evaluate", map[string]any{}, adminCookies)
	requireStatus(t, evaluateRec, http.StatusOK)

	evaluatedGetRec := app.get(t, "/api/v1/campaigns/"+evalCampaignID, adminCookies)
	requireStatus(t, evaluatedGetRec, http.StatusOK)
	evaluatedCampaign := decodeSuccess[map[string]any](t, evaluatedGetRec)
	if evaluatedCampaign["evaluated_at"] == nil {
		t.Fatalf("expected evaluated_at after evaluation, got %#v", evaluatedCampaign)
	}

	cancelRec := app.post(t, "/api/v1/campaigns/"+cancelCampaignID+"/cancel", map[string]any{}, adminCookies)
	requireStatus(t, cancelRec, http.StatusOK)

	forbiddenCancel := app.post(t, "/api/v1/campaigns/"+cancelCampaignID+"/cancel", map[string]any{}, app.login(t, member))
	requireStatus(t, forbiddenCancel, http.StatusForbidden)
}

func TestCampaigns_CancelAndEvaluateRejectInvalidCampaignID(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	adminCookies := app.login(t, admin)

	badCancelRec := app.post(t, "/api/v1/campaigns/not-a-uuid/cancel", map[string]any{}, adminCookies)
	requireStatus(t, badCancelRec, http.StatusBadRequest)

	badEvaluateRec := app.post(t, "/api/v1/campaigns/not-a-uuid/evaluate", map[string]any{}, adminCookies)
	requireStatus(t, badEvaluateRec, http.StatusBadRequest)
}
