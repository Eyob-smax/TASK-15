package api_tests

import (
	"net/http"
	"strings"
	"testing"
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
