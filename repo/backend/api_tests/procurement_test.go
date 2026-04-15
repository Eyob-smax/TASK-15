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

	approvedGetRec := app.get(t, "/api/v1/purchase-orders/"+poID, procurementCookies)
	requireStatus(t, approvedGetRec, http.StatusOK)
	approvedPO := decodeSuccess[map[string]any](t, approvedGetRec)
	if approvedPO["approved_by"] == nil || approvedPO["approved_at"] == nil {
		t.Fatalf("expected approved_by and approved_at after approval, got %#v", approvedPO)
	}

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

	receivedGetRec := app.get(t, "/api/v1/purchase-orders/"+poID, procurementCookies)
	requireStatus(t, receivedGetRec, http.StatusOK)
	receivedPO := decodeSuccess[map[string]any](t, receivedGetRec)
	if receivedPO["received_at"] == nil {
		t.Fatalf("expected received_at after receipt, got %#v", receivedPO)
	}

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

	getVarianceRec := app.get(t, "/api/v1/variances/"+shortageID, procurementCookies)
	requireStatus(t, getVarianceRec, http.StatusOK)

	resolveRec := app.post(t, "/api/v1/variances/"+shortageID+"/resolve", map[string]any{
		"action":           "adjustment",
		"resolution_notes": "Stock corrected after recount",
		"quantity_change":  2,
	}, procurementCookies)
	requireStatus(t, resolveRec, http.StatusOK)

	resolvedVarianceRec := app.get(t, "/api/v1/variances/"+shortageID, procurementCookies)
	requireStatus(t, resolvedVarianceRec, http.StatusOK)
	resolvedVariance := decodeSuccess[map[string]any](t, resolvedVarianceRec)
	if resolvedVariance["resolved_at"] == nil {
		t.Fatalf("expected resolved_at after resolution, got %#v", resolvedVariance)
	}

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

func TestProcurement_SupplierListGetAndUpdate(t *testing.T) {
	app := newIntegrationApp(t)
	procurement := app.seedUser(t, "procurement_specialist", nil)
	cookies := app.login(t, procurement)

	createRec := app.post(t, "/api/v1/suppliers", map[string]any{
		"name":          "Listable Supplier",
		"contact_name":  "One",
		"contact_email": "one@fitcommerce.test",
		"contact_phone": "555-1212",
		"address":       "10 Supplier Lane",
	}, cookies)
	requireStatus(t, createRec, http.StatusCreated)
	created := decodeSuccess[map[string]any](t, createRec)
	id := created["id"].(string)

	listRec := app.get(t, "/api/v1/suppliers", cookies)
	requireStatus(t, listRec, http.StatusOK)
	rows, _ := decodePaginated[map[string]any](t, listRec)
	if len(rows) == 0 {
		t.Fatal("expected supplier list to contain at least one supplier")
	}

	getRec := app.get(t, "/api/v1/suppliers/"+id, cookies)
	requireStatus(t, getRec, http.StatusOK)

	updateRec := app.put(t, "/api/v1/suppliers/"+id, map[string]any{
		"name": "Updated Supplier Name",
	}, cookies)
	requireStatus(t, updateRec, http.StatusOK)
	updated := decodeSuccess[map[string]any](t, updateRec)
	if updated["name"] != "Updated Supplier Name" {
		t.Fatalf("expected updated supplier name, got %#v", updated["name"])
	}
}

func TestProcurement_POListGetReturnVoidAndLandedCostSummary(t *testing.T) {
	app := newIntegrationApp(t)
	admin := app.seedUser(t, "administrator", nil)
	procurement := app.seedUser(t, "procurement_specialist", nil)
	cookies := app.login(t, procurement)

	item := app.seedItem(t, itemSeedOptions{
		CreatedBy: admin.ID,
		Name:      "Returnable Item",
		Status:    "published",
		Quantity:  5,
	})

	supplierID := app.seedSupplier(t, "Lifecycle Supplier")

	createRec := app.post(t, "/api/v1/purchase-orders", map[string]any{
		"supplier_id": supplierID.String(),
		"lines": []map[string]any{{
			"item_id":            item.ID.String(),
			"ordered_quantity":   6,
			"ordered_unit_price": 9.25,
		}},
	}, cookies)
	requireStatus(t, createRec, http.StatusCreated)
	po := decodeSuccess[map[string]any](t, createRec)
	poID := po["id"].(string)
	line := po["lines"].([]any)[0].(map[string]any)

	listRec := app.get(t, "/api/v1/purchase-orders", cookies)
	requireStatus(t, listRec, http.StatusOK)
	poRows, _ := decodePaginated[map[string]any](t, listRec)
	if len(poRows) == 0 {
		t.Fatal("expected purchase-order list to contain at least one record")
	}

	getRec := app.get(t, "/api/v1/purchase-orders/"+poID, cookies)
	requireStatus(t, getRec, http.StatusOK)

	approveRec := app.post(t, "/api/v1/purchase-orders/"+poID+"/approve", map[string]any{}, cookies)
	requireStatus(t, approveRec, http.StatusOK)

	receiveRec := app.post(t, "/api/v1/purchase-orders/"+poID+"/receive", map[string]any{
		"lines": []map[string]any{{
			"po_line_id":          line["id"].(string),
			"received_quantity":   4,
			"received_unit_price": 9.25,
		}},
	}, cookies)
	requireStatus(t, receiveRec, http.StatusOK)

	returnRec := app.post(t, "/api/v1/purchase-orders/"+poID+"/return", map[string]any{}, cookies)
	requireStatus(t, returnRec, http.StatusOK)

	voidPORec := app.post(t, "/api/v1/purchase-orders", map[string]any{
		"supplier_id": supplierID.String(),
		"lines": []map[string]any{{
			"item_id":            item.ID.String(),
			"ordered_quantity":   1,
			"ordered_unit_price": 9.25,
		}},
	}, cookies)
	requireStatus(t, voidPORec, http.StatusCreated)
	voidPOID := decodeSuccess[map[string]any](t, voidPORec)["id"].(string)

	approveVoidPORec := app.post(t, "/api/v1/purchase-orders/"+voidPOID+"/approve", map[string]any{}, cookies)
	requireStatus(t, approveVoidPORec, http.StatusOK)

	voidRec := app.post(t, "/api/v1/purchase-orders/"+voidPOID+"/void", map[string]any{}, cookies)
	requireStatus(t, voidRec, http.StatusOK)

	landedSummaryRec := app.get(t, "/api/v1/procurement/landed-costs?item_id="+item.ID.String()+"&period=current", cookies)
	requireStatus(t, landedSummaryRec, http.StatusOK)
	_ = decodeSuccess[[]map[string]any](t, landedSummaryRec)
}

func TestProcurement_InvalidIDsAndMissingQueryParamsReturnBadRequest(t *testing.T) {
	app := newIntegrationApp(t)
	procurement := app.seedUser(t, "procurement_specialist", nil)
	cookies := app.login(t, procurement)

	badSupplierGet := app.get(t, "/api/v1/suppliers/not-a-uuid", cookies)
	requireStatus(t, badSupplierGet, http.StatusBadRequest)

	badSupplierUpdate := app.put(t, "/api/v1/suppliers/not-a-uuid", map[string]any{"name": "x"}, cookies)
	requireStatus(t, badSupplierUpdate, http.StatusBadRequest)

	badPOGet := app.get(t, "/api/v1/purchase-orders/not-a-uuid", cookies)
	requireStatus(t, badPOGet, http.StatusBadRequest)

	badPOReturn := app.post(t, "/api/v1/purchase-orders/not-a-uuid/return", map[string]any{}, cookies)
	requireStatus(t, badPOReturn, http.StatusBadRequest)

	badPOVoid := app.post(t, "/api/v1/purchase-orders/not-a-uuid/void", map[string]any{}, cookies)
	requireStatus(t, badPOVoid, http.StatusBadRequest)

	missingItemID := app.get(t, "/api/v1/procurement/landed-costs", cookies)
	requireStatus(t, missingItemID, http.StatusBadRequest)

	badItemID := app.get(t, "/api/v1/procurement/landed-costs?item_id=bad", cookies)
	requireStatus(t, badItemID, http.StatusBadRequest)
}
