import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient, type PaginatedResponse } from '@/lib/api-client';
import { enqueueOfflineMutation } from '@/lib/offline-cache';
import type { InventorySnapshot, InventoryAdjustment, WarehouseBin } from '@/lib/types';

interface SnapshotParams {
  item_id?: string;
  location_id?: string;
}

export function useInventorySnapshots(params: SnapshotParams = {}) {
  return useQuery({
    queryKey: ['inventory', 'snapshots', params],
    queryFn: () =>
      apiClient.get<{ data: InventorySnapshot[] }>('/inventory/snapshots', {
        item_id: params.item_id,
        location_id: params.location_id,
      }),
  });
}

interface AdjustmentListParams {
  item_id?: string;
  page?: number;
  page_size?: number;
}

export function useInventoryAdjustments(params: AdjustmentListParams = {}) {
  return useQuery({
    queryKey: ['inventory', 'adjustments', params],
    queryFn: () =>
      apiClient.get<PaginatedResponse<InventoryAdjustment>>('/inventory/adjustments', {
        item_id: params.item_id,
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
      }),
  });
}

interface CreateAdjustmentPayload {
  item_id: string;
  quantity_change: number;
  reason: string;
}

export function useCreateAdjustment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: CreateAdjustmentPayload) => {
      if (typeof navigator !== 'undefined' && !navigator.onLine) {
        await enqueueOfflineMutation({
          id: globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-create-adjustment`,
          type: 'create-adjustment',
          payload: payload as unknown as Record<string, unknown>,
          createdAt: Date.now(),
        });
        return;
      }
      return apiClient.post<{ data: InventoryAdjustment }>('/inventory/adjustments', payload);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['inventory'] });
    },
  });
}

interface WarehouseBinListParams {
  location_id?: string;
  page?: number;
  page_size?: number;
}

export function useWarehouseBins(params: WarehouseBinListParams = {}) {
  return useQuery({
    queryKey: ['warehouse-bins', params],
    queryFn: () =>
      apiClient.get<PaginatedResponse<WarehouseBin>>('/warehouse-bins', {
        location_id: params.location_id,
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
      }),
  });
}

export function useWarehouseBin(id: string | undefined) {
  return useQuery({
    queryKey: ['warehouse-bins', id],
    queryFn: async () => {
      const response = await apiClient.get<{ data: WarehouseBin }>(`/warehouse-bins/${id}`);
      return response.data;
    },
    enabled: Boolean(id),
  });
}
