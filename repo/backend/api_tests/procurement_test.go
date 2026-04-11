package api_tests

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestProcurement_MemberCannotCreateSupplier(t *testing.T) {
	app := newIntegrationApp(t)
	member := app.seedUser(t, "member", nil)

	rec := app.post(t, "/api/v1/suppliers", map[string]any{
		"name": "Blocked Supplier",
	}, app.login(t, member))
	requireStatus(t, rec, http.StatusForbidden)
}

func TestProcurement_CreateApproveReceiveAndResolveVarianceFlow(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	procurement := app.seedUser(t, "procurement_specialist", nil)
	procurementCookies := app.login(t, procurement)
	item := app.seedItem(t, itemSeedOptions{
		CreatedBy: admin.ID,
		Name:      "Protein Tub",
		Status:    "published",
		Quantity:  3,
	})

	supplierRec := app.post(t, "/api/v1/suppliers", map[string]any{
		"name":          "Supplement Source",
		"contact_name":  "Vendor One",
		"contact_email": "vendor.one@fitcommerce.test",
		"contact_phone": "555-2000",
		"address":       "42 Vendor Road",
	}, procurementCookies)
	requireStatus(t, supplierRec, http.StatusCreated)
	supplier := decodeSuccess[map[string]any](t, supplierRec)

	createPORec := app.post(t, "/api/v1/purchase-orders", map[string]any{
		"supplier_id": supplier["id"].(string),
		"lines": []map[string]any{
			{
				"item_id":             item.ID.String(),
				"ordered_quantity":    10,
				"ordered_unit_price":  7.50,
			},
		},
	}, procurementCookies)
	requireStatus(t, createPORec, http.StatusCreated)
	po := decodeSuccess[map[string]any](t, createPORec)
	poID := po["id"].(string)
	lines := po["lines"].([]any)
	if len(lines) != 1 {
		t.Fatalf("expected one PO line, got %#v", po["lines"])
	}
	line := lines[0].(map[string]any)
	if int(line["ordered_quantity"].(float64)) != 10 {
		t.Fatalf("expected ordered_quantity 10, got %#v", line["ordered_quantity"])
	}

	approveRec := app.post(t, "/api/v1/purchase-orders/"+poID+"/approve", map[string]any{}, procurementCookies)
	requireStatus(t, approveRec, http.StatusOK)

	receiveRec := app.post(t, "/api/v1/purchase-orders/"+poID+"/receive", map[string]any{
		"lines": []map[string]any{
			{
				"po_line_id":            line["id"].(string),
				"received_quantity":     8,
				"received_unit_price":   8.25,
			},
		},
	}, procurementCookies)
	requireStatus(t, receiveRec, http.StatusOK)

	variancesRec := app.get(t, "/api/v1/variances", procurementCookies)
	requireStatus(t, variancesRec, http.StatusOK)
	variances, _ := decodePaginated[map[string]any](t, variancesRec)
	if len(variances) < 2 {
		t.Fatalf("expected quantity and price variances, got %#v", variances)
	}

	var shortageID string
	for _, variance := range variances {
		if variance["type"] == "shortage" {
			shortageID = variance["id"].(string)
			break
		}
	}
	if shortageID == "" {
		t.Fatal("expected a shortage variance to resolve")
	}

	resolveRec := app.post(t, "/api/v1/variances/"+shortageID+"/resolve", map[string]any{
		"action":           "adjustment",
		"resolution_notes": "Stock corrected after recount",
		"quantity_change":  2,
	}, procurementCookies)
	requireStatus(t, resolveRec, http.StatusOK)

	landedCostRec := app.get(t, "/api/v1/procurement/landed-costs/"+poID, procurementCookies)
	requireStatus(t, landedCostRec, http.StatusOK)
	entries := decodeSuccess[[]map[string]any](t, landedCostRec)
	if len(entries) == 0 {
		t.Fatal("expected landed cost entries after PO receipt")
	}
}

func TestProcurement_EscalatedVarianceCanBeResolved(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	procurement := app.seedUser(t, "procurement_specialist", nil)
	procurementCookies := app.login(t, procurement)
	item := app.seedItem(t, itemSeedOptions{
		CreatedBy: admin.ID,
		Name:      "Creatine Bundle",
		Status:    "published",
		Quantity:  4,
	})

	supplierID := app.seedSupplier(t, "Escalation Supplier")

	createPORec := app.post(t, "/api/v1/purchase-orders", map[string]any{
		"supplier_id": supplierID.String(),
		"lines": []map[string]any{
			{
				"item_id":            item.ID.String(),
				"ordered_quantity":   10,
				"ordered_unit_price": 12.50,
			},
		},
	}, procurementCookies)
	requireStatus(t, createPORec, http.StatusCreated)

	po := decodeSuccess[map[string]any](t, createPORec)
	poID := po["id"].(string)
	line := po["lines"].([]any)[0].(map[string]any)

	approveRec := app.post(t, "/api/v1/purchase-orders/"+poID+"/approve", map[string]any{}, procurementCookies)
	requireStatus(t, approveRec, http.StatusOK)

	receiveRec := app.post(t, "/api/v1/purchase-orders/"+poID+"/receive", map[string]any{
		"lines": []map[string]any{
			{
				"po_line_id":          line["id"].(string),
				"received_quantity":   7,
				"received_unit_price": 12.50,
			},
		},
	}, procurementCookies)
	requireStatus(t, receiveRec, http.StatusOK)

	variancesRec := app.get(t, "/api/v1/variances?status=open", procurementCookies)
	requireStatus(t, variancesRec, http.StatusOK)
	variances, _ := decodePaginated[map[string]any](t, variancesRec)
	if len(variances) == 0 {
		t.Fatal("expected at least one open variance")
	}

	var varianceID string
	for _, variance := range variances {
		if variance["type"] == "shortage" {
			varianceID = variance["id"].(string)
			break
		}
	}
	if varianceID == "" {
		t.Fatal("expected a shortage variance to escalate")
	}

	if _, err := app.app.Pool.Exec(
		context.Background(),
		`UPDATE variance_records SET status = 'escalated', resolution_due_date = $2 WHERE id = $1`,
		varianceID,
		time.Now().UTC().Add(-24*time.Hour),
	); err != nil {
		t.Fatalf("mark variance escalated: %v", err)
	}

	resolveRec := app.post(t, "/api/v1/variances/"+varianceID+"/resolve", map[string]any{
		"action":           "adjustment",
		"resolution_notes": "Escalated shortage resolved after final inventory reconciliation",
		"quantity_change":  3,
	}, procurementCookies)
	requireStatus(t, resolveRec, http.StatusOK)
}
