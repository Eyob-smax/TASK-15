package api_tests

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestReports_CoachCannotReadAdminOnlyReportData(t *testing.T) {
	app := newIntegrationApp(t)
	coach := app.seedUser(t, "coach", nil)
	memberGrowthID := app.reportIDByType(t, "member_growth")

	rec := app.get(t, "/api/v1/reports/"+memberGrowthID.String()+"/data", app.login(t, coach))
	requireStatus(t, rec, http.StatusForbidden)

	if decodeError(t, rec).Error.Code != "FORBIDDEN" {
		t.Fatal("expected FORBIDDEN for coach on admin-only report")
	}
}

func TestReports_CoachWithoutLocationIsForbiddenFromData(t *testing.T) {
	app := newIntegrationApp(t)
	// Coach with no assigned location — passes role check for engagement report
	// but must be rejected by the location scope enforcement.
	coachNoLocation := app.seedUser(t, "coach", nil)
	engagementID := app.reportIDByType(t, "engagement")

	rec := app.get(t, "/api/v1/reports/"+engagementID.String()+"/data", app.login(t, coachNoLocation))
	requireStatus(t, rec, http.StatusForbidden)

	if decodeError(t, rec).Error.Code != "FORBIDDEN" {
		t.Fatal("expected FORBIDDEN for coach without assigned location")
	}
}

func TestReports_ExportDownloadUsesRealFilesAndAccessControl(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	coach := app.seedUser(t, "coach", nil)
	app.seedItem(t, itemSeedOptions{
		CreatedBy: admin.ID,
		Name:      "Export Item",
		Status:    "published",
		Quantity:  7,
	})

	reportID := app.reportIDByType(t, "inventory_summary")
	adminCookies := app.login(t, admin)

	exportRec := app.post(t, "/api/v1/exports", map[string]any{
		"report_id": reportID.String(),
		"format":    "csv",
		"parameters": map[string]string{
			"category": "strength",
		},
	}, adminCookies)
	requireStatus(t, exportRec, http.StatusAccepted)

	exportJob := decodeSuccess[map[string]any](t, exportRec)
	exportID := exportJob["id"].(string)

	getRec := app.get(t, "/api/v1/exports/"+exportID, adminCookies)
	requireStatus(t, getRec, http.StatusOK)

	downloadRec := app.get(t, "/api/v1/exports/"+exportID+"/download", adminCookies)
	requireStatus(t, downloadRec, http.StatusOK)
	if !strings.Contains(downloadRec.Header().Get("Content-Disposition"), "attachment") {
		t.Fatalf("expected attachment content disposition, got %q", downloadRec.Header().Get("Content-Disposition"))
	}
	if len(downloadRec.Body.Bytes()) == 0 {
		t.Fatal("expected export download body to contain file contents")
	}

	coachRec := app.get(t, "/api/v1/exports/"+exportID, app.login(t, coach))
	requireStatus(t, coachRec, http.StatusForbidden)
}

func TestReports_CoachWithLocationCannotReadOtherLocationData(t *testing.T) {
	app := newIntegrationApp(t)

	locA := app.seedLocation(t, "Gym A")
	locB := app.seedLocation(t, "Gym B")

	// Coach is assigned to location A.
	coachA := app.seedUser(t, "coach", &locA.ID)

	// Admin seeds a published item belonging to location B.
	admin := app.seedUser(t, "administrator", nil)
	itemB := app.seedItem(t, itemSeedOptions{
		CreatedBy:  admin.ID,
		LocationID: &locB.ID,
		Name:       "Location B Kettlebell",
		Status:     "published",
		Quantity:   10,
		AvailabilityWindows: []timeWindow{{
			Start: time.Now().UTC().Add(-1 * time.Hour),
			End:   time.Now().UTC().Add(48 * time.Hour),
		}},
	})

	// A member creates an order for the location-B item (seeds engagement data at B).
	memberB := app.seedUser(t, "member", &locB.ID)
	orderRec := app.post(t, "/api/v1/orders", map[string]any{
		"item_id":  itemB.ID.String(),
		"quantity": 1,
	}, app.login(t, memberB))
	requireStatus(t, orderRec, http.StatusCreated)

	// Admin creates a campaign for the location-B item (seeds class_fill_rate data at B).
	campaignRec := app.post(t, "/api/v1/campaigns", map[string]any{
		"item_id":      itemB.ID.String(),
		"min_quantity": 5,
		"cutoff_time":  time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	}, app.login(t, admin))
	requireStatus(t, campaignRec, http.StatusCreated)

	coachCookies := app.login(t, coachA)

	// Coach A requests engagement report — must see 0 records (B's orders excluded).
	engagementID := app.reportIDByType(t, "engagement")
	engRec := app.get(t, "/api/v1/reports/"+engagementID.String()+"/data", coachCookies)
	requireStatus(t, engRec, http.StatusOK)
	engResult := decodeSuccess[map[string]any](t, engRec)
	engRecords, _ := engResult["records"].([]any)
	if len(engRecords) != 0 {
		t.Fatalf("expected coach at location A to see 0 engagement records from location B, got %d", len(engRecords))
	}

	// Coach A requests class_fill_rate report — must see 0 records (B's campaigns excluded).
	classFillID := app.reportIDByType(t, "class_fill_rate")
	cfRec := app.get(t, "/api/v1/reports/"+classFillID.String()+"/data", coachCookies)
	requireStatus(t, cfRec, http.StatusOK)
	cfResult := decodeSuccess[map[string]any](t, cfRec)
	cfRecords, _ := cfResult["records"].([]any)
	if len(cfRecords) != 0 {
		t.Fatalf("expected coach at location A to see 0 class_fill_rate records from location B, got %d", len(cfRecords))
	}
}

func TestReports_CoachExportIsLocationScoped(t *testing.T) {
	app := newIntegrationApp(t)

	locA := app.seedLocation(t, "Export Gym A")
	locB := app.seedLocation(t, "Export Gym B")

	coachA := app.seedUser(t, "coach", &locA.ID)

	admin := app.seedUser(t, "administrator", nil)
	itemB := app.seedItem(t, itemSeedOptions{
		CreatedBy:  admin.ID,
		LocationID: &locB.ID,
		Name:       "Export Location B Item",
		Status:     "published",
		Quantity:   10,
		AvailabilityWindows: []timeWindow{{
			Start: time.Now().UTC().Add(-1 * time.Hour),
			End:   time.Now().UTC().Add(48 * time.Hour),
		}},
	})

	// Seed an order for the location-B item so it would appear in engagement export.
	memberB := app.seedUser(t, "member", &locB.ID)
	orderRec := app.post(t, "/api/v1/orders", map[string]any{
		"item_id":  itemB.ID.String(),
		"quantity": 1,
	}, app.login(t, memberB))
	requireStatus(t, orderRec, http.StatusCreated)
	locBOrder := decodeSuccess[map[string]any](t, orderRec)
	locBOrderID := locBOrder["id"].(string)

	coachCookies := app.login(t, coachA)

	// Coach A generates a CSV export of the engagement report.
	engagementID := app.reportIDByType(t, "engagement")
	exportRec := app.post(t, "/api/v1/exports", map[string]any{
		"report_id": engagementID.String(),
		"format":    "csv",
	}, coachCookies)
	requireStatus(t, exportRec, http.StatusAccepted)
	exportJob := decodeSuccess[map[string]any](t, exportRec)
	exportID := exportJob["id"].(string)

	// Export is generated synchronously — download immediately.
	downloadRec := app.get(t, "/api/v1/exports/"+exportID+"/download", coachCookies)
	requireStatus(t, downloadRec, http.StatusOK)

	// The location-B order ID must not appear in the exported content.
	csvBody := downloadRec.Body.String()
	if strings.Contains(csvBody, locBOrderID) {
		t.Fatalf("export for coach at location A must not contain order from location B (%s)", locBOrderID)
	}
}
