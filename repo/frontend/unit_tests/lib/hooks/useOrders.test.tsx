import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import {
  useOrderList,
  useOrder,
  useOrderTimeline,
  useCancelOrder,
  usePayOrder,
  useRefundOrder,
  useAddOrderNote,
  useSplitOrder,
  useMergeOrder,
} from '@/lib/hooks/useOrders';

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

describe('useOrderList', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches with default page/page_size', async () => {
    mockGet.mockResolvedValue({ data: [], total: 0 });
    const { result } = renderHook(() => useOrderList(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockGet).toHaveBeenCalledWith('/orders', {
      page: 1,
      page_size: 20,
      status: undefined,
    });
  });

  it('passes custom params', async () => {
    mockGet.mockResolvedValue({ data: [], total: 0 });
    renderHook(() => useOrderList({ page: 3, page_size: 50, status: 'paid' }), { wrapper });
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith('/orders', {
      page: 3,
      page_size: 50,
      status: 'paid',
    });
  });
});

describe('useOrder / useOrderTimeline', () => {
  beforeEach(() => vi.clearAllMocks());

  it('useOrder disabled without id', () => {
    const { result } = renderHook(() => useOrder(undefined), { wrapper });
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('useOrder fetches single order', async () => {
    mockGet.mockResolvedValue({ data: { id: 'o1' } });
    const { result } = renderHook(() => useOrder('o1'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockGet).toHaveBeenCalledWith('/orders/o1');
    expect(result.current.data).toEqual({ id: 'o1' });
  });

  it('useOrderTimeline fetches timeline', async () => {
    mockGet.mockResolvedValue({ data: [{ id: 't1' }] });
    const { result } = renderHook(() => useOrderTimeline('o1'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockGet).toHaveBeenCalledWith('/orders/o1/timeline');
  });
});

describe('useCancelOrder', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setOnline(true);
  });

  it('posts to /cancel when online', async () => {
    mockPost.mockResolvedValue({ data: { ok: true } });
    const { result } = renderHook(() => useCancelOrder(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync('o1');
    });
    expect(mockPost).toHaveBeenCalledWith('/orders/o1/cancel', {});
  });

  it('enqueues offline mutation when offline', async () => {
    setOnline(false);
    const { result } = renderHook(() => useCancelOrder(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync('o1');
    });
    expect(mockEnqueueOffline).toHaveBeenCalled();
    expect(mockPost).not.toHaveBeenCalled();
  });
});

describe('usePayOrder', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setOnline(true);
  });

  it('posts payment with settlement_marker', async () => {
    mockPost.mockResolvedValue({ data: { ok: true } });
    const { result } = renderHook(() => usePayOrder(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ id: 'o1', settlementMarker: 'sm-42' });
    });
    expect(mockPost).toHaveBeenCalledWith('/orders/o1/pay', { settlement_marker: 'sm-42' });
  });

  it('enqueues offline when offline', async () => {
    setOnline(false);
    const { result } = renderHook(() => usePayOrder(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ id: 'o1', settlementMarker: 'sm-42' });
    });
    expect(mockEnqueueOffline).toHaveBeenCalled();
  });
});

describe('useRefundOrder / useAddOrderNote', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setOnline(true);
  });

  it('useRefundOrder posts to /refund', async () => {
    mockPost.mockResolvedValue({ data: { ok: true } });
    const { result } = renderHook(() => useRefundOrder(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync('o1');
    });
    expect(mockPost).toHaveBeenCalledWith('/orders/o1/refund', {});
  });

  it('useAddOrderNote posts note', async () => {
    mockPost.mockResolvedValue({ data: { ok: true } });
    const { result } = renderHook(() => useAddOrderNote(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ id: 'o1', note: 'hello' });
    });
    expect(mockPost).toHaveBeenCalledWith('/orders/o1/notes', { note: 'hello' });
  });
});

describe('useSplitOrder / useMergeOrder', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setOnline(true);
  });

  it('useSplitOrder posts split body', async () => {
    mockPost.mockResolvedValue({ data: [{ id: 'o1a' }, { id: 'o1b' }] });
    const { result } = renderHook(() => useSplitOrder(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ id: 'o1', quantities: [3, 2] });
    });
    expect(mockPost).toHaveBeenCalledWith(
      '/orders/o1/split',
      expect.objectContaining({ quantities: [3, 2] }),
    );
  });

  it('useMergeOrder posts merge body', async () => {
    mockPost.mockResolvedValue({ data: { id: 'om' } });
    const { result } = renderHook(() => useMergeOrder(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ order_ids: ['o1', 'o2'] });
    });
    expect(mockPost).toHaveBeenCalledWith(
      '/orders/merge',
      expect.objectContaining({ order_ids: ['o1', 'o2'] }),
    );
  });

  it('useSplitOrder enqueues offline', async () => {
    setOnline(false);
    const { result } = renderHook(() => useSplitOrder(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ id: 'o1', quantities: [1, 1] });
    });
    expect(mockEnqueueOffline).toHaveBeenCalled();
    expect(mockPost).not.toHaveBeenCalled();
  });
});
