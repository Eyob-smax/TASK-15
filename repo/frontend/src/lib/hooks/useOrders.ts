import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient, type ApiEnvelope, type ApiMessage, type PaginatedResponse } from '@/lib/api-client';
import { enqueueOfflineMutation } from '@/lib/offline-cache';
import type { Order, OrderTimelineEntry } from '@/lib/types';

interface OrderListParams {
  page?: number;
  page_size?: number;
  status?: string;
}

export function useOrderList(params: OrderListParams = {}) {
  return useQuery({
    queryKey: ['orders', params],
    queryFn: () =>
      apiClient.get<PaginatedResponse<Order>>('/orders', {
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
        status: params.status,
      }),
  });
}

export function useOrder(id: string | undefined) {
  return useQuery({
    queryKey: ['orders', id],
    queryFn: async () => {
      const response = await apiClient.get<ApiEnvelope<Order>>(`/orders/${id}`);
      return response.data;
    },
    enabled: Boolean(id),
  });
}

export function useOrderTimeline(id: string | undefined) {
  return useQuery({
    queryKey: ['orders', id, 'timeline'],
    queryFn: async () => {
      const response = await apiClient.get<ApiEnvelope<OrderTimelineEntry[]>>(`/orders/${id}/timeline`);
      return response.data;
    },
    enabled: Boolean(id),
  });
}

export function useCancelOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      if (typeof navigator !== 'undefined' && !navigator.onLine) {
        await enqueueOfflineMutation({
          id: globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-cancel-order`,
          type: 'cancel-order',
          payload: { id },
          createdAt: Date.now(),
        });
        return;
      }
      return apiClient.post<ApiEnvelope<ApiMessage>>(`/orders/${id}/cancel`, {});
    },
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: ['orders'] });
      qc.invalidateQueries({ queryKey: ['orders', id] });
    },
  });
}

interface PayOrderPayload {
  id: string;
  settlementMarker: string;
}

export function usePayOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, settlementMarker }: PayOrderPayload) => {
      if (typeof navigator !== 'undefined' && !navigator.onLine) {
        await enqueueOfflineMutation({
          id: globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-pay-order`,
          type: 'pay-order',
          payload: { id, settlementMarker },
          createdAt: Date.now(),
        });
        return;
      }
      return apiClient.post<ApiEnvelope<ApiMessage>>(`/orders/${id}/pay`, {
        settlement_marker: settlementMarker,
      });
    },
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ['orders'] });
      qc.invalidateQueries({ queryKey: ['orders', id] });
    },
  });
}

export function useRefundOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      if (typeof navigator !== 'undefined' && !navigator.onLine) {
        await enqueueOfflineMutation({
          id: globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-refund-order`,
          type: 'refund-order',
          payload: { id },
          createdAt: Date.now(),
        });
        return;
      }
      return apiClient.post<ApiEnvelope<ApiMessage>>(`/orders/${id}/refund`, {});
    },
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: ['orders'] });
      qc.invalidateQueries({ queryKey: ['orders', id] });
    },
  });
}

interface AddNotePayload {
  id: string;
  note: string;
}

export function useAddOrderNote() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, note }: AddNotePayload) => {
      if (typeof navigator !== 'undefined' && !navigator.onLine) {
        await enqueueOfflineMutation({
          id: globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-add-order-note`,
          type: 'add-order-note',
          payload: { id, note },
          createdAt: Date.now(),
        });
        return;
      }
      return apiClient.post(`/orders/${id}/notes`, { note });
    },
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ['orders', id, 'timeline'] });
    },
  });
}

export function useSplitOrder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      quantities,
      supplier_id,
      warehouse_bin_id,
      pickup_point,
    }: {
      id: string;
      quantities: number[];
      supplier_id?: string;
      warehouse_bin_id?: string;
      pickup_point?: string;
    }) => {
      if (typeof navigator !== 'undefined' && !navigator.onLine) {
        await enqueueOfflineMutation({
          id: globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-split-order`,
          type: 'split-order',
          payload: { id, quantities, supplier_id, warehouse_bin_id, pickup_point },
          createdAt: Date.now(),
        });
        return;
      }
      const response = await apiClient.post<ApiEnvelope<Order[]>>(`/orders/${id}/split`, {
        quantities,
        supplier_id,
        warehouse_bin_id,
        pickup_point,
      });
      return response.data;
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
      queryClient.invalidateQueries({ queryKey: ['orders', variables.id] });
      queryClient.invalidateQueries({ queryKey: ['orders', variables.id, 'timeline'] });
    },
  });
}

export function useMergeOrder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({
      order_ids,
      supplier_id,
      warehouse_bin_id,
      pickup_point,
    }: {
      order_ids: string[];
      supplier_id?: string;
      warehouse_bin_id?: string;
      pickup_point?: string;
    }) => {
      if (typeof navigator !== 'undefined' && !navigator.onLine) {
        await enqueueOfflineMutation({
          id: globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-merge-order`,
          type: 'merge-order',
          payload: { order_ids, supplier_id, warehouse_bin_id, pickup_point },
          createdAt: Date.now(),
        });
        return;
      }
      const response = await apiClient.post<ApiEnvelope<Order>>(`/orders/merge`, {
        order_ids,
        supplier_id,
        warehouse_bin_id,
        pickup_point,
      });
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
    },
  });
}
