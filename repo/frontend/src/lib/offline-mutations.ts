import type { QueryClient } from "@tanstack/react-query";
import { apiClient, isOfflineApiError } from "@/lib/api-client";
import {
  loadPendingOfflineMutations,
  markOfflineMutationFailed,
  removeOfflineMutation,
  setOfflineItemIDMapping,
  type OfflineMutationEntry,
} from "@/lib/offline-cache";

function isBrowserOffline(): boolean {
  return typeof navigator !== "undefined" && !navigator.onLine;
}

async function executeMutation(entry: OfflineMutationEntry): Promise<void> {
  switch (entry.type) {
    case "create-item": {
      const temporaryID =
        typeof entry.payload.__offline_temp_id === "string"
          ? entry.payload.__offline_temp_id
          : undefined;
      const payload = { ...entry.payload };
      delete payload.__offline_temp_id;

      const response = await apiClient.post<
        Record<string, unknown> | { data?: Record<string, unknown> }
      >("/items", payload);

      const responseRecord = response as Record<string, unknown>;
      const topLevelID = responseRecord["id"];
      const nestedData = responseRecord["data"];

      const serverID =
        (typeof topLevelID === "string" ? topLevelID : undefined) ??
        (typeof nestedData === "object" &&
        nestedData &&
        typeof (nestedData as Record<string, unknown>).id === "string"
          ? ((nestedData as Record<string, unknown>).id as string)
          : undefined);

      if (temporaryID && serverID) {
        await setOfflineItemIDMapping(temporaryID, serverID);
      }
      return;
    }
    case "update-item": {
      const id = entry.payload.id;
      const body = entry.payload.body;
      if (typeof id !== "string" || !body || typeof body !== "object") {
        return;
      }
      await apiClient.put(`/items/${id}`, body);
      return;
    }
    case "run-export":
      await apiClient.post("/exports", entry.payload);
      return;
    case "cancel-order":
      await apiClient.post(`/orders/${entry.payload.id}/cancel`, {});
      return;
    case "pay-order":
      await apiClient.post(`/orders/${entry.payload.id}/pay`, {
        settlement_marker: entry.payload.settlementMarker,
      });
      return;
    case "join-campaign":
      await apiClient.post(`/campaigns/${entry.payload.id}/join`, {
        quantity: entry.payload.quantity,
      });
      return;
    case "cancel-campaign":
      await apiClient.post(`/campaigns/${entry.payload.id}/cancel`, {});
      return;
    case "create-po":
      await apiClient.post("/purchase-orders", entry.payload);
      return;
    case "approve-po":
      await apiClient.post(`/purchase-orders/${entry.payload.id}/approve`, {});
      return;
    case "receive-po":
      await apiClient.post(
        `/purchase-orders/${entry.payload.id}/receive`,
        entry.payload.body,
      );
      return;
    case "void-po":
      await apiClient.post(`/purchase-orders/${entry.payload.id}/void`, {});
      return;
    case "refund-order":
      await apiClient.post(`/orders/${entry.payload.id}/refund`, {});
      return;
    case "add-order-note":
      await apiClient.post(`/orders/${entry.payload.id}/notes`, {
        note: entry.payload.note,
      });
      return;
    case "split-order":
      await apiClient.post(`/orders/${entry.payload.id}/split`, {
        quantities: entry.payload.quantities,
        supplier_id: entry.payload.supplier_id,
        warehouse_bin_id: entry.payload.warehouse_bin_id,
        pickup_point: entry.payload.pickup_point,
      });
      return;
    case "merge-order":
      await apiClient.post(`/orders/merge`, {
        order_ids: entry.payload.order_ids,
        supplier_id: entry.payload.supplier_id,
        warehouse_bin_id: entry.payload.warehouse_bin_id,
        pickup_point: entry.payload.pickup_point,
      });
      return;
    case "evaluate-campaign":
      await apiClient.post(`/campaigns/${entry.payload.id}/evaluate`, {});
      return;
    case "create-campaign":
      await apiClient.post("/campaigns", entry.payload);
      return;
    case "create-adjustment":
      await apiClient.post("/inventory/adjustments", entry.payload);
      return;
    case "return-po":
      await apiClient.post(`/purchase-orders/${entry.payload.id}/return`, {});
      return;
    case "resolve-variance":
      await apiClient.post(`/variances/${entry.payload.id}/resolve`, {
        action: entry.payload.action,
        resolution_notes: entry.payload.resolution_notes,
        quantity_change: entry.payload.quantity_change,
      });
      return;
    default:
      return;
  }
}

export async function replayOfflineMutations(
  queryClient: QueryClient,
): Promise<number> {
  if (isBrowserOffline()) {
    return 0;
  }

  const queue = await loadPendingOfflineMutations();
  let processed = 0;

  for (const entry of queue) {
    if (entry.status === "failed") {
      continue;
    }

    try {
      await executeMutation(entry);
      await removeOfflineMutation(entry.id);
      processed += 1;
    } catch (error) {
      if (isOfflineApiError(error)) {
        break;
      }

      const message =
        typeof error === "object" && error && "message" in error
          ? String((error as { message?: unknown }).message ?? "replay failed")
          : "replay failed";
      await markOfflineMutationFailed(entry.id, message);
    }
  }

  if (processed > 0) {
    queryClient.invalidateQueries({ queryKey: ["items"] });
    queryClient.invalidateQueries({ queryKey: ["orders"] });
    queryClient.invalidateQueries({ queryKey: ["campaigns"] });
    queryClient.invalidateQueries({ queryKey: ["purchase-orders"] });
    queryClient.invalidateQueries({ queryKey: ["inventory"] });
    queryClient.invalidateQueries({ queryKey: ["variances"] });
    queryClient.invalidateQueries({ queryKey: ["reports"] });
    queryClient.invalidateQueries({ queryKey: ["exports"] });
  }

  return processed;
}
