import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  apiClient,
  type ApiEnvelope,
  type ApiMessage,
  type PaginatedResponse,
} from "@/lib/api-client";
import {
  enqueueOfflineMutation,
  resolveOfflineItemID,
} from "@/lib/offline-cache";
import type { BatchEditResponse, Item } from "@/lib/types";

interface ItemFilters {
  category?: string;
  brand?: string;
  condition?: string;
  status?: string;
}

interface ItemListParams extends ItemFilters {
  page?: number;
  page_size?: number;
}

export function useItemList(params: ItemListParams = {}) {
  return useQuery({
    queryKey: ["items", params],
    queryFn: () =>
      apiClient.get<PaginatedResponse<Item>>("/items", {
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
        category: params.category,
        brand: params.brand,
        condition: params.condition,
        status: params.status,
      }),
  });
}

export function useItem(id: string | undefined) {
  return useQuery({
    queryKey: ["items", id],
    queryFn: async () => {
      const resolvedID = await resolveOfflineItemID(id as string);
      const response = await apiClient.get<ApiEnvelope<Item>>(
        `/items/${resolvedID}`,
      );
      return response.data;
    },
    enabled: Boolean(id),
  });
}

export function usePublishItem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      apiClient.post<ApiEnvelope<ApiMessage>>(`/items/${id}/publish`, {}),
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: ["items"] });
      qc.invalidateQueries({ queryKey: ["items", id] });
    },
  });
}

export function useUnpublishItem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      apiClient.post<ApiEnvelope<ApiMessage>>(`/items/${id}/unpublish`, {}),
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: ["items"] });
      qc.invalidateQueries({ queryKey: ["items", id] });
    },
  });
}

interface UpdateItemPayload {
  id: string;
  body: Record<string, unknown>;
}

export function useUpdateItem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, body }: UpdateItemPayload) => {
      if (typeof navigator !== "undefined" && !navigator.onLine) {
        await enqueueOfflineMutation({
          id: globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-update-item`,
          type: "update-item",
          payload: { id, body },
          createdAt: Date.now(),
        });
        return { id, ...(body as Record<string, unknown>) } as Item;
      }
      const response = await apiClient.put<ApiEnvelope<Item>>(
        `/items/${id}`,
        body,
      );
      return response.data;
    },
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ["items"] });
      qc.invalidateQueries({ queryKey: ["items", id] });
    },
  });
}

export interface BatchEditWindowInput {
  start_time: string;
  end_time: string;
}

export interface BatchEditRow {
  item_id: string;
  field: string;
  new_value?: string;
  availability_windows?: BatchEditWindowInput[];
}

export function useBatchEdit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (edits: BatchEditRow[]) => {
      const response = await apiClient.post<ApiEnvelope<BatchEditResponse>>(
        "/items/batch-edit",
        { edits },
      );
      return response.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["items"] });
    },
  });
}

export function useCreateItem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: Record<string, unknown>) => {
      if (typeof navigator !== "undefined" && !navigator.onLine) {
        const queuedID =
          globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-create-item`;
        await enqueueOfflineMutation({
          id: queuedID,
          type: "create-item",
          payload: { ...body, __offline_temp_id: queuedID },
          createdAt: Date.now(),
        });
        return { id: queuedID, ...(body as Record<string, unknown>) } as Item;
      }
      const response = await apiClient.post<ApiEnvelope<Item>>("/items", body);
      return response.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["items"] });
    },
  });
}
