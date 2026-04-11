import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient, type PaginatedResponse } from '@/lib/api-client';
import { enqueueOfflineMutation } from '@/lib/offline-cache';
import type { GroupBuyCampaign, GroupBuyParticipant } from '@/lib/types';

interface CampaignListParams {
  page?: number;
  page_size?: number;
  status?: string;
}

interface SingleCampaignResponse {
  data: GroupBuyCampaign;
}

export function useCampaignList(params: CampaignListParams = {}) {
  return useQuery({
    queryKey: ['campaigns', params],
    queryFn: () =>
      apiClient.get<PaginatedResponse<GroupBuyCampaign>>('/campaigns', {
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
        status: params.status,
      }),
  });
}

export function useCampaign(id: string | undefined) {
  return useQuery({
    queryKey: ['campaigns', id],
    queryFn: async () => {
      const response = await apiClient.get<SingleCampaignResponse>(`/campaigns/${id}`);
      return response.data;
    },
    enabled: Boolean(id),
  });
}

interface JoinCampaignPayload {
  id: string;
  quantity: number;
}

interface CreateCampaignPayload {
  item_id: string;
  min_quantity: number;
  cutoff_time: string;
}

export function useJoinCampaign() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, quantity }: JoinCampaignPayload) => {
      if (typeof navigator !== 'undefined' && !navigator.onLine) {
        await enqueueOfflineMutation({
          id: globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-join-campaign`,
          type: 'join-campaign',
          payload: { id, quantity },
          createdAt: Date.now(),
        });
        return;
      }
      return apiClient.post<{ data: GroupBuyParticipant }>(`/campaigns/${id}/join`, { quantity });
    },
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ['campaigns'] });
      qc.invalidateQueries({ queryKey: ['campaigns', id] });
      qc.invalidateQueries({ queryKey: ['orders'] });
    },
  });
}

export function useCancelCampaign() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      if (typeof navigator !== 'undefined' && !navigator.onLine) {
        await enqueueOfflineMutation({
          id: globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-cancel-campaign`,
          type: 'cancel-campaign',
          payload: { id },
          createdAt: Date.now(),
        });
        return;
      }
      return apiClient.post<SingleCampaignResponse>(`/campaigns/${id}/cancel`, {});
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['campaigns'] });
    },
  });
}

export function useEvaluateCampaign() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      if (typeof navigator !== 'undefined' && !navigator.onLine) {
        await enqueueOfflineMutation({
          id: globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-evaluate-campaign`,
          type: 'evaluate-campaign',
          payload: { id },
          createdAt: Date.now(),
        });
        return;
      }
      return apiClient.post<SingleCampaignResponse>(`/campaigns/${id}/evaluate`, {});
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['campaigns'] });
    },
  });
}

export function useCreateCampaign() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: CreateCampaignPayload) => {
      if (typeof navigator !== 'undefined' && !navigator.onLine) {
        await enqueueOfflineMutation({
          id: globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-create-campaign`,
          type: 'create-campaign',
          payload: body as unknown as Record<string, unknown>,
          createdAt: Date.now(),
        });
        return;
      }
      return apiClient.post<SingleCampaignResponse>('/campaigns', body);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['campaigns'] });
    },
  });
}
