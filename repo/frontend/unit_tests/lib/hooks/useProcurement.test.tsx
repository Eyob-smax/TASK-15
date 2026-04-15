import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import {
  useSupplierList,
  useSupplier,
  useCreateSupplier,
  useUpdateSupplier,
  usePOList,
  usePO,
  useCreatePO,
  useApprovePO,
  useReceivePO,
  useReturnPO,
  useVoidPO,
  useVarianceList,
  useVariance,
  useResolveVariance,
} from '@/lib/hooks/useProcurement';

const mockGet = vi.fn();
const mockPost = vi.fn();
const mockPut = vi.fn();
const mockEnqueueOffline = vi.fn();

vi.mock('@/lib/api-client', () => ({
  apiClient: {
    get: (...args: unknown[]) => mockGet(...args),
    post: (...args: unknown[]) => mockPost(...args),
    put: (...args: unknown[]) => mockPut(...args),
  },
}));

vi.mock('@/lib/offline-cache', () => ({
  enqueueOfflineMutation: (...args: unknown[]) => mockEnqueueOffline(...args),
}));

function wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

function setOnline(value: boolean) {
  Object.defineProperty(window.navigator, 'onLine', {
    configurable: true,
    value,
  });
}

describe('suppliers', () => {
  beforeEach(() => vi.clearAllMocks());

  it('useSupplierList fetches with defaults', async () => {
    mockGet.mockResolvedValue({ data: [] });
    renderHook(() => useSupplierList(), { wrapper });
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith(
      '/suppliers',
      expect.objectContaining({ page: 1, page_size: 20 }),
    );
  });

  it('useSupplierList forwards search', async () => {
    mockGet.mockResolvedValue({ data: [] });
    renderHook(() => useSupplierList({ search: 'acme' }), { wrapper });
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith(
      '/suppliers',
      expect.objectContaining({ search: 'acme' }),
    );
  });

  it('useSupplier disabled without id', () => {
    const { result } = renderHook(() => useSupplier(undefined), { wrapper });
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('useSupplier fetches single and unwraps', async () => {
    mockGet.mockResolvedValue({ data: { id: 's1' } });
    const { result } = renderHook(() => useSupplier('s1'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockGet).toHaveBeenCalledWith('/suppliers/s1');
    expect(result.current.data).toEqual({ id: 's1' });
  });

  it('useCreateSupplier posts and returns data', async () => {
    mockPost.mockResolvedValue({ data: { id: 's1', name: 'Acme' } });
    const { result } = renderHook(() => useCreateSupplier(), { wrapper });
    let returned: unknown;
    await act(async () => {
      returned = await result.current.mutateAsync({ name: 'Acme' });
    });
    expect(mockPost).toHaveBeenCalledWith('/suppliers', { name: 'Acme' });
    expect(returned).toEqual({ id: 's1', name: 'Acme' });
  });

  it('useUpdateSupplier puts and returns data', async () => {
    mockPut.mockResolvedValue({ data: { id: 's1', is_active: false } });
    const { result } = renderHook(() => useUpdateSupplier(), { wrapper });
    let returned: unknown;
    await act(async () => {
      returned = await result.current.mutateAsync({ id: 's1', body: { is_active: false } });
    });
    expect(mockPut).toHaveBeenCalledWith('/suppliers/s1', { is_active: false });
    expect(returned).toMatchObject({ id: 's1' });
  });
});

describe('purchase orders', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setOnline(true);
  });

  it('usePOList fetches with defaults', async () => {
    mockGet.mockResolvedValue({ data: [] });
    renderHook(() => usePOList(), { wrapper });
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith(
      '/purchase-orders',
      expect.objectContaining({ page: 1, page_size: 20 }),
    );
  });

  it('usePOList forwards status and supplier_id', async () => {
    mockGet.mockResolvedValue({ data: [] });
    renderHook(() => usePOList({ status: 'approved', supplier_id: 's1' }), { wrapper });
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith(
      '/purchase-orders',
      expect.objectContaining({ status: 'approved', supplier_id: 's1' }),
    );
  });

  it('usePO disabled without id', () => {
    const { result } = renderHook(() => usePO(undefined), { wrapper });
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('usePO fetches single and unwraps', async () => {
    mockGet.mockResolvedValue({ data: { id: 'po-1' } });
    const { result } = renderHook(() => usePO('po-1'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockGet).toHaveBeenCalledWith('/purchase-orders/po-1');
    expect(result.current.data).toEqual({ id: 'po-1' });
  });

  it('useCreatePO posts when online', async () => {
    mockPost.mockResolvedValue({ data: { id: 'po-1' } });
    const { result } = renderHook(() => useCreatePO(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({
        supplier_id: 's1',
        lines: [{ item_id: 'i1', ordered_quantity: 10, ordered_unit_price: 5 }],
      });
    });
    expect(mockPost).toHaveBeenCalledWith(
      '/purchase-orders',
      expect.objectContaining({ supplier_id: 's1' }),
    );
  });

  it('useCreatePO enqueues when offline', async () => {
    setOnline(false);
    const { result } = renderHook(() => useCreatePO(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({
        supplier_id: 's1',
        lines: [{ item_id: 'i1', ordered_quantity: 1, ordered_unit_price: 1 }],
      });
    });
    expect(mockEnqueueOffline).toHaveBeenCalled();
    expect(mockPost).not.toHaveBeenCalled();
  });

  it('useApprovePO posts when online and enqueues when offline', async () => {
    mockPost.mockResolvedValue({ data: {} });
    const { result } = renderHook(() => useApprovePO(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync('po-1');
    });
    expect(mockPost).toHaveBeenCalledWith('/purchase-orders/po-1/approve', {});

    setOnline(false);
    vi.clearAllMocks();
    const { result: result2 } = renderHook(() => useApprovePO(), { wrapper });
    await act(async () => {
      await result2.current.mutateAsync('po-1');
    });
    expect(mockEnqueueOffline).toHaveBeenCalled();
    expect(mockPost).not.toHaveBeenCalled();
  });

  it('useReceivePO posts lines when online', async () => {
    mockPost.mockResolvedValue({ data: {} });
    const { result } = renderHook(() => useReceivePO(), { wrapper });
    const lines = [{ po_line_id: 'l1', received_quantity: 3, received_unit_price: 5 }];
    await act(async () => {
      await result.current.mutateAsync({ id: 'po-1', lines });
    });
    expect(mockPost).toHaveBeenCalledWith('/purchase-orders/po-1/receive', { lines });
  });

  it('useReceivePO enqueues when offline', async () => {
    setOnline(false);
    const { result } = renderHook(() => useReceivePO(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ id: 'po-1', lines: [] });
    });
    expect(mockEnqueueOffline).toHaveBeenCalled();
    expect(mockPost).not.toHaveBeenCalled();
  });

  it('useReturnPO posts when online', async () => {
    mockPost.mockResolvedValue({ data: {} });
    const { result } = renderHook(() => useReturnPO(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync('po-1');
    });
    expect(mockPost).toHaveBeenCalledWith('/purchase-orders/po-1/return', {});
  });

  it('useVoidPO posts when online and enqueues when offline', async () => {
    mockPost.mockResolvedValue({ data: {} });
    const { result } = renderHook(() => useVoidPO(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync('po-1');
    });
    expect(mockPost).toHaveBeenCalledWith('/purchase-orders/po-1/void', {});

    setOnline(false);
    vi.clearAllMocks();
    const { result: result2 } = renderHook(() => useVoidPO(), { wrapper });
    await act(async () => {
      await result2.current.mutateAsync('po-1');
    });
    expect(mockEnqueueOffline).toHaveBeenCalled();
  });
});

describe('variances', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setOnline(true);
  });

  it('useVarianceList fetches with defaults', async () => {
    mockGet.mockResolvedValue({ data: [] });
    renderHook(() => useVarianceList(), { wrapper });
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith(
      '/variances',
      expect.objectContaining({ page: 1, page_size: 20 }),
    );
  });

  it('useVarianceList forwards filters', async () => {
    mockGet.mockResolvedValue({ data: [] });
    renderHook(
      () => useVarianceList({ status: 'open', purchase_order_id: 'po-1' }),
      { wrapper },
    );
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith(
      '/variances',
      expect.objectContaining({ status: 'open', purchase_order_id: 'po-1' }),
    );
  });

  it('useVariance disabled without id', () => {
    const { result } = renderHook(() => useVariance(undefined), { wrapper });
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('useVariance fetches single and unwraps', async () => {
    mockGet.mockResolvedValue({ data: { id: 'v1' } });
    const { result } = renderHook(() => useVariance('v1'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockGet).toHaveBeenCalledWith('/variances/v1');
    expect(result.current.data).toEqual({ id: 'v1' });
  });

  it('useResolveVariance posts when online', async () => {
    mockPost.mockResolvedValue({ data: { id: 'v1', status: 'resolved' } });
    const { result } = renderHook(() => useResolveVariance(), { wrapper });
    let returned: unknown;
    await act(async () => {
      returned = await result.current.mutateAsync({
        id: 'v1',
        action: 'adjustment',
        resolution_notes: 'done',
        quantity_change: 2,
      });
    });
    expect(mockPost).toHaveBeenCalledWith('/variances/v1/resolve', {
      action: 'adjustment',
      resolution_notes: 'done',
      quantity_change: 2,
    });
    expect(returned).toMatchObject({ id: 'v1' });
  });

  it('useResolveVariance enqueues when offline', async () => {
    setOnline(false);
    const { result } = renderHook(() => useResolveVariance(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({
        id: 'v1',
        action: 'return',
        resolution_notes: 'n',
      });
    });
    expect(mockEnqueueOffline).toHaveBeenCalled();
    expect(mockPost).not.toHaveBeenCalled();
  });
});
