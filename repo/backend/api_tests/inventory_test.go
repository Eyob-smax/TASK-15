package api_tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestInventory_RoleAccessAndSnapshotListing(t *testing.T) {
	app := newIntegrationApp(t)
	location := app.seedLocation(t, "Warehouse East")
	admin := app.seedUser(t, "administrator", &location.ID)
	procurement := app.seedUser(t, "procurement_specialist", &location.ID)
	member := app.seedUser(t, "member", nil)
	item := app.seedItem(t, itemSeedOptions{
		CreatedBy: admin.ID,
		Name:      "Inventory Snapshot Item",
		Status:    "published",
		Quantity:  12,
		LocationID: &location.ID,
		AvailabilityWindows: []timeWindow{{
			Start: time.Now().UTC().Add(-2 * time.Hour),
			End:   time.Now().UTC().Add(24 * time.Hour),
		}},
	})

	_, err := app.app.Pool.Exec(context.Background(),
		`INSERT INTO inventory_snapshots (id, item_id, quantity, location_id, recorded_at)
		 VALUES ($1, $2, $3, $4, NOW())`,
		uuid.New(), item.ID, item.Quantity, location.ID,
	)
	if err != nil {
		t.Fatalf("seed inventory snapshot: %v", err)
	}

	memberRec := app.get(t, "/api/v1/inventory/snapshots", app.login(t, member))
	requireStatus(t, memberRec, http.StatusForbidden)

	procurementRec := app.get(t, "/api/v1/inventory/snapshots?item_id="+item.ID.String(), app.login(t, procurement))
	requireStatus(t, procurementRec, http.StatusOK)

	snapshots := decodeSuccess[[]map[string]any](t, procurementRec)
	if len(snapshots) != 1 {
		t.Fatalf("expected exactly one inventory snapshot, got %d", len(snapshots))
	}
}

func TestInventory_CreateAdjustmentAndWarehouseBin(t *testing.T) {
	app := newIntegrationApp(t)
	location := app.seedLocation(t, "Warehouse West")
	ops := app.seedUser(t, "operations_manager", &location.ID)
	item := app.seedItem(t, itemSeedOptions{
		CreatedBy: ops.ID,
		Name:      "Adjustable Dumbbell",
		Status:    "published",
		Quantity:  6,
		LocationID: &location.ID,
		AvailabilityWindows: []timeWindow{{
			Start: time.Now().UTC().Add(-2 * time.Hour),
			End:   time.Now().UTC().Add(24 * time.Hour),
		}},
	})
	opsCookies := app.login(t, ops)

	binRec := app.post(t, "/api/v1/warehouse-bins", map[string]any{
		"location_id":  location.ID.String(),
		"name":         "BIN-A1",
		"description":  "Primary rack bin",
	}, opsCookies)
	requireStatus(t, binRec, http.StatusCreated)
	binID := decodeSuccess[map[string]any](t, binRec)["id"].(string)

	adjustRec := app.post(t, "/api/v1/inventory/adjustments", map[string]any{
		"item_id":         item.ID.String(),
		"quantity_change": 4,
		"reason":          "manual recount",
	}, opsCookies)
	requireStatus(t, adjustRec, http.StatusCreated)

	listAdjustmentsRec := app.get(t, "/api/v1/inventory/adjustments?item_id="+item.ID.String(), opsCookies)
	requireStatus(t, listAdjustmentsRec, http.StatusOK)
	adjustments, _ := decodePaginated[map[string]any](t, listAdjustmentsRec)
	if len(adjustments) == 0 {
		t.Fatal("expected at least one inventory adjustment")
	}

	listBinsRec := app.get(t, "/api/v1/warehouse-bins?location_id="+location.ID.String(), opsCookies)
	requireStatus(t, listBinsRec, http.StatusOK)
	bins, _ := decodePaginated[map[string]any](t, listBinsRec)
	if len(bins) != 1 {
		t.Fatalf("expected one warehouse bin, got %d", len(bins))
	}

	getBinRec := app.get(t, "/api/v1/warehouse-bins/"+binID, opsCookies)
	requireStatus(t, getBinRec, http.StatusOK)
	bin := decodeSuccess[map[string]any](t, getBinRec)
	if bin["name"] != "BIN-A1" {
		t.Fatalf("expected BIN-A1 from bin detail endpoint, got %#v", bin["name"])
	}
}
