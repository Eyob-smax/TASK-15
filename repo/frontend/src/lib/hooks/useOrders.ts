import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient, type ApiEnvelope, type ApiMessage, type PaginatedResponse } from '@/lib/api-client';
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
    mutationFn: (id: string) =>
      apiClient.post<ApiEnvelope<ApiMessage>>(`/orders/${id}/cancel`, {}),
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
    mutationFn: ({ id, settlementMarker }: PayOrderPayload) =>
      apiClient.post<ApiEnvelope<ApiMessage>>(`/orders/${id}/pay`, {
        settlement_marker: settlementMarker,
      }),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ['orders'] });
      qc.invalidateQueries({ queryKey: ['orders', id] });
    },
  });
}

export function useRefundOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      apiClient.post<ApiEnvelope<ApiMessage>>(`/orders/${id}/refund`, {}),
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
    mutationFn: ({ id, note }: AddNotePayload) =>
      apiClient.post(`/orders/${id}/notes`, { note }),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ['orders', id, 'timeline'] });
    },
  });
}

export function useSplitOrder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, quantities }: { id: string; quantities: number[] }) => {
      const response = await apiClient.post<ApiEnvelope<Order[]>>(`/orders/${id}/split`, { quantities });
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
    mutationFn: async ({ order_ids }: { order_ids: string[] }) => {
      const response = await apiClient.post<ApiEnvelope<Order>>(`/orders/merge`, { order_ids });
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
    },
  });
}
