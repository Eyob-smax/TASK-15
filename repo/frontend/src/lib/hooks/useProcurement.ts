import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient, type ApiEnvelope, type ApiMessage, type PaginatedResponse } from '@/lib/api-client';
import type { Supplier, PurchaseOrder, VarianceRecord } from '@/lib/types';

// ─── Shared ──────────────────────────────────────────────────────────────────

// ─── Supplier param / body types ─────────────────────────────────────────────

interface SupplierListParams {
  page?: number;
  page_size?: number;
  search?: string;
}

interface CreateSupplierBody {
  name: string;
  contact_name?: string;
  contact_email?: string;
  contact_phone?: string;
  address?: string;
  payment_terms?: string;
  lead_time_days?: number;
  notes?: string;
}

interface UpdateSupplierPayload {
  id: string;
  body: Partial<CreateSupplierBody> & { is_active?: boolean };
}

// ─── PO param / body types ───────────────────────────────────────────────────

interface POListParams {
  page?: number;
  page_size?: number;
  status?: string;
  supplier_id?: string;
}

interface POLineBody {
  item_id: string;
  ordered_quantity: number;
  ordered_unit_price: number;
}

interface CreatePOBody {
  supplier_id: string;
  lines: POLineBody[];
  order_date?: string;
  expected_delivery_date?: string;
  currency?: string;
  shipping_cost?: number;
  insurance_cost?: number;
  customs_duty?: number;
  other_costs?: number;
  notes?: string;
}

interface ReceivePOLine {
  po_line_id: string;
  received_quantity: number;
  received_unit_price: number;
}

interface ReceivePOPayload {
  id: string;
  lines: ReceivePOLine[];
}

// ─── Variance param / body types ─────────────────────────────────────────────

interface VarianceListParams {
  page?: number;
  page_size?: number;
  status?: string;
  purchase_order_id?: string;
}

interface ResolveVariancePayload {
  id: string;
  action: 'adjustment' | 'return';
  resolution_notes: string;
  quantity_change?: number;
}

// ─── Supplier hooks ───────────────────────────────────────────────────────────

export function useSupplierList(params: SupplierListParams = {}) {
  return useQuery({
    queryKey: ['suppliers', params],
    queryFn: () =>
      apiClient.get<PaginatedResponse<Supplier>>('/suppliers', {
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
        search: params.search,
      }),
  });
}

export function useSupplier(id: string | undefined) {
  return useQuery({
    queryKey: ['suppliers', id],
    queryFn: async () => {
      const response = await apiClient.get<ApiEnvelope<Supplier>>(`/suppliers/${id}`);
      return response.data;
    },
    enabled: Boolean(id),
  });
}

export function useCreateSupplier() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: CreateSupplierBody) => {
      const response = await apiClient.post<ApiEnvelope<Supplier>>('/suppliers', body);
      return response.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['suppliers'] });
    },
  });
}

export function useUpdateSupplier() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, body }: UpdateSupplierPayload) => {
      const response = await apiClient.put<ApiEnvelope<Supplier>>(`/suppliers/${id}`, body);
      return response.data;
    },
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ['suppliers'] });
      qc.invalidateQueries({ queryKey: ['suppliers', id] });
    },
  });
}

// ─── Purchase Order hooks ─────────────────────────────────────────────────────

export function usePOList(params: POListParams = {}) {
  return useQuery({
    queryKey: ['purchase-orders', params],
    queryFn: () =>
      apiClient.get<PaginatedResponse<PurchaseOrder>>('/purchase-orders', {
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
        status: params.status,
        supplier_id: params.supplier_id,
      }),
  });
}

export function usePO(id: string | undefined) {
  return useQuery({
    queryKey: ['purchase-orders', id],
    queryFn: async () => {
      const response = await apiClient.get<ApiEnvelope<PurchaseOrder>>(`/purchase-orders/${id}`);
      return response.data;
    },
    enabled: Boolean(id),
  });
}

export function useCreatePO() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: CreatePOBody) => {
      const response = await apiClient.post<ApiEnvelope<PurchaseOrder>>('/purchase-orders', body);
      return response.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['purchase-orders'] });
    },
  });
}

export function useApprovePO() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      apiClient.post<ApiEnvelope<ApiMessage>>(`/purchase-orders/${id}/approve`, {}),
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: ['purchase-orders'] });
      qc.invalidateQueries({ queryKey: ['purchase-orders', id] });
    },
  });
}

export function useReceivePO() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, lines }: ReceivePOPayload) =>
      apiClient.post<ApiEnvelope<ApiMessage>>(`/purchase-orders/${id}/receive`, { lines }),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ['purchase-orders'] });
      qc.invalidateQueries({ queryKey: ['purchase-orders', id] });
      qc.invalidateQueries({ queryKey: ['variances'] });
    },
  });
}

export function useReturnPO() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      apiClient.post<ApiEnvelope<ApiMessage>>(`/purchase-orders/${id}/return`, {}),
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: ['purchase-orders'] });
      qc.invalidateQueries({ queryKey: ['purchase-orders', id] });
    },
  });
}

export function useVoidPO() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      apiClient.post<ApiEnvelope<ApiMessage>>(`/purchase-orders/${id}/void`, {}),
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: ['purchase-orders'] });
      qc.invalidateQueries({ queryKey: ['purchase-orders', id] });
    },
  });
}

// ─── Variance hooks ───────────────────────────────────────────────────────────

export function useVarianceList(params: VarianceListParams = {}) {
  return useQuery({
    queryKey: ['variances', params],
    queryFn: () =>
      apiClient.get<PaginatedResponse<VarianceRecord>>('/variances', {
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
        status: params.status,
        purchase_order_id: params.purchase_order_id,
      }),
  });
}

export function useVariance(id: string | undefined) {
  return useQuery({
    queryKey: ['variances', id],
    queryFn: async () => {
      const response = await apiClient.get<ApiEnvelope<VarianceRecord>>(`/variances/${id}`);
      return response.data;
    },
    enabled: Boolean(id),
  });
}

export function useResolveVariance() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, action, resolution_notes, quantity_change }: ResolveVariancePayload) => {
      const response = await apiClient.post<ApiEnvelope<VarianceRecord>>(`/variances/${id}/resolve`, {
        action,
        resolution_notes,
        quantity_change,
      });
      return response.data;
    },
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ['variances'] });
      qc.invalidateQueries({ queryKey: ['variances', id] });
    },
  });
}

// Re-export PurchaseOrderLine type usage for consumers
export type { ReceivePOLine, CreateSupplierBody, CreatePOBody, POLineBody };
