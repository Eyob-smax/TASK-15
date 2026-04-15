import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import {
  useItemList,
  useItem,
  usePublishItem,
  useUnpublishItem,
  useUpdateItem,
  useBatchEdit,
  useCreateItem,
} from '@/lib/hooks/useItems';

const mockGet = vi.fn();
const mockPost = vi.fn();
const mockPut = vi.fn();
const mockEnqueueOffline = vi.fn();
const mockResolveOfflineID = vi.fn();

vi.mock('@/lib/api-client', () => ({
  apiClient: {
    get: (...args: unknown[]) => mockGet(...args),
    post: (...args: unknown[]) => mockPost(...args),
    put: (...args: unknown[]) => mockPut(...args),
  },
}));

vi.mock('@/lib/offline-cache', () => ({
  enqueueOfflineMutation: (...args: unknown[]) => mockEnqueueOffline(...args),
  resolveOfflineItemID: (...args: unknown[]) => mockResolveOfflineID(...args),
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

describe('useItemList', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls /items with defaults', async () => {
    mockGet.mockResolvedValue({ data: [], total: 0 });
    renderHook(() => useItemList(), { wrapper });
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith('/items', expect.objectContaining({ page: 1, page_size: 20 }));
  });

  it('forwards filter params', async () => {
    mockGet.mockResolvedValue({ data: [] });
    renderHook(
      () => useItemList({ category: 'cardio', brand: 'Acme', condition: 'new', status: 'published' }),
      { wrapper },
    );
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith(
      '/items',
      expect.objectContaining({
        category: 'cardio',
        brand: 'Acme',
        condition: 'new',
        status: 'published',
      }),
    );
  });
});

describe('useItem', () => {
  beforeEach(() => vi.clearAllMocks());

  it('is idle when id is undefined', () => {
    const { result } = renderHook(() => useItem(undefined), { wrapper });
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('resolves offline id and fetches', async () => {
    mockResolveOfflineID.mockResolvedValue('real-id');
    mockGet.mockResolvedValue({ data: { id: 'real-id' } });

    const { result } = renderHook(() => useItem('temp-id'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockResolveOfflineID).toHaveBeenCalledWith('temp-id');
    expect(mockGet).toHaveBeenCalledWith('/items/real-id');
    expect(result.current.data).toEqual({ id: 'real-id' });
  });
});

describe('usePublishItem / useUnpublishItem', () => {
  beforeEach(() => vi.clearAllMocks());

  it('usePublishItem posts to publish endpoint', async () => {
    mockPost.mockResolvedValue({ data: { ok: true } });
    const { result } = renderHook(() => usePublishItem(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync('item-1');
    });
    expect(mockPost).toHaveBeenCalledWith('/items/item-1/publish', {});
  });

  it('useUnpublishItem posts to unpublish endpoint', async () => {
    mockPost.mockResolvedValue({ data: { ok: true } });
    const { result } = renderHook(() => useUnpublishItem(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync('item-1');
    });
    expect(mockPost).toHaveBeenCalledWith('/items/item-1/unpublish', {});
  });
});

describe('useUpdateItem', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setOnline(true);
  });

  it('puts when online', async () => {
    mockPut.mockResolvedValue({ data: { id: 'item-1', name: 'new' } });
    const { result } = renderHook(() => useUpdateItem(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ id: 'item-1', body: { name: 'new' } });
    });
    expect(mockPut).toHaveBeenCalledWith('/items/item-1', { name: 'new' });
  });

  it('enqueues when offline', async () => {
    setOnline(false);
    const { result } = renderHook(() => useUpdateItem(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ id: 'item-1', body: { name: 'x' } });
    });
    expect(mockEnqueueOffline).toHaveBeenCalled();
    expect(mockPut).not.toHaveBeenCalled();
  });
});

describe('useCreateItem', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setOnline(true);
  });

  it('posts when online', async () => {
    mockPost.mockResolvedValue({ data: { id: 'new-id' } });
    const { result } = renderHook(() => useCreateItem(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ name: 'Bike' });
    });
    expect(mockPost).toHaveBeenCalledWith('/items', { name: 'Bike' });
  });

  it('enqueues when offline and returns temporary record', async () => {
    setOnline(false);
    const { result } = renderHook(() => useCreateItem(), { wrapper });
    let returned: unknown;
    await act(async () => {
      returned = await result.current.mutateAsync({ name: 'Bike' });
    });
    expect(mockEnqueueOffline).toHaveBeenCalled();
    expect(returned).toMatchObject({ name: 'Bike' });
  });
});

describe('useBatchEdit', () => {
  beforeEach(() => vi.clearAllMocks());

  it('posts edits array', async () => {
    mockPost.mockResolvedValue({ data: { updated: 2, failures: [] } });
    const { result } = renderHook(() => useBatchEdit(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync([
        { item_id: 'i1', field: 'status', new_value: 'published' },
      ]);
    });
    expect(mockPost).toHaveBeenCalledWith(
      '/items/batch-edit',
      expect.objectContaining({ edits: expect.any(Array) }),
    );
  });
});
