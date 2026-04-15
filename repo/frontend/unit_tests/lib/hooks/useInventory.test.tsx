import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import {
  useInventorySnapshots,
  useInventoryAdjustments,
  useCreateAdjustment,
  useWarehouseBins,
  useWarehouseBin,
} from '@/lib/hooks/useInventory';

const mockGet = vi.fn();
const mockPost = vi.fn();
const mockEnqueueOffline = vi.fn();

vi.mock('@/lib/api-client', () => ({
  apiClient: {
    get: (...args: unknown[]) => mockGet(...args),
    post: (...args: unknown[]) => mockPost(...args),
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

describe('useInventorySnapshots', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches snapshots with filters', async () => {
    mockGet.mockResolvedValue({ data: [] });
    renderHook(
      () => useInventorySnapshots({ item_id: 'i1', location_id: 'loc-1' }),
      { wrapper },
    );
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith('/inventory/snapshots', {
      item_id: 'i1',
      location_id: 'loc-1',
    });
  });

  it('fetches snapshots with no filters', async () => {
    mockGet.mockResolvedValue({ data: [] });
    renderHook(() => useInventorySnapshots(), { wrapper });
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith('/inventory/snapshots', {
      item_id: undefined,
      location_id: undefined,
    });
  });
});

describe('useInventoryAdjustments', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches with defaults', async () => {
    mockGet.mockResolvedValue({ data: [], total: 0 });
    renderHook(() => useInventoryAdjustments(), { wrapper });
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith(
      '/inventory/adjustments',
      expect.objectContaining({ page: 1, page_size: 20 }),
    );
  });

  it('passes item_id and pagination', async () => {
    mockGet.mockResolvedValue({ data: [] });
    renderHook(() => useInventoryAdjustments({ item_id: 'i1', page: 2, page_size: 50 }), {
      wrapper,
    });
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith('/inventory/adjustments', {
      item_id: 'i1',
      page: 2,
      page_size: 50,
    });
  });
});

describe('useCreateAdjustment', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setOnline(true);
  });

  it('posts when online', async () => {
    mockPost.mockResolvedValue({ data: { id: 'adj-1' } });
    const { result } = renderHook(() => useCreateAdjustment(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({
        item_id: 'i1',
        quantity_change: -3,
        reason: 'damaged',
      });
    });
    expect(mockPost).toHaveBeenCalledWith('/inventory/adjustments', {
      item_id: 'i1',
      quantity_change: -3,
      reason: 'damaged',
    });
  });

  it('enqueues offline', async () => {
    setOnline(false);
    const { result } = renderHook(() => useCreateAdjustment(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ item_id: 'i1', quantity_change: 1, reason: 'restock' });
    });
    expect(mockEnqueueOffline).toHaveBeenCalled();
    expect(mockPost).not.toHaveBeenCalled();
  });
});

describe('useWarehouseBins / useWarehouseBin', () => {
  beforeEach(() => vi.clearAllMocks());

  it('useWarehouseBins applies defaults', async () => {
    mockGet.mockResolvedValue({ data: [] });
    renderHook(() => useWarehouseBins(), { wrapper });
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith(
      '/warehouse-bins',
      expect.objectContaining({ page: 1, page_size: 20 }),
    );
  });

  it('useWarehouseBins forwards location_id', async () => {
    mockGet.mockResolvedValue({ data: [] });
    renderHook(() => useWarehouseBins({ location_id: 'loc-1' }), { wrapper });
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith(
      '/warehouse-bins',
      expect.objectContaining({ location_id: 'loc-1' }),
    );
  });

  it('useWarehouseBin disabled without id', () => {
    const { result } = renderHook(() => useWarehouseBin(undefined), { wrapper });
    expect(result.current.fetchStatus).toBe('idle');
    expect(mockGet).not.toHaveBeenCalled();
  });

  it('useWarehouseBin fetches by id and unwraps data', async () => {
    mockGet.mockResolvedValue({ data: { id: 'bin-1' } });
    const { result } = renderHook(() => useWarehouseBin('bin-1'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockGet).toHaveBeenCalledWith('/warehouse-bins/bin-1');
    expect(result.current.data).toEqual({ id: 'bin-1' });
  });
});
