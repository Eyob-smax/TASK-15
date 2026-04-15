import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import {
  useCampaignList,
  useCampaign,
  useJoinCampaign,
  useCancelCampaign,
  useEvaluateCampaign,
  useCreateCampaign,
} from '@/lib/hooks/useCampaigns';

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

describe('useCampaignList', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches with defaults', async () => {
    mockGet.mockResolvedValue({ data: [] });
    renderHook(() => useCampaignList(), { wrapper });
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith(
      '/campaigns',
      expect.objectContaining({ page: 1, page_size: 20 }),
    );
  });

  it('forwards status filter', async () => {
    mockGet.mockResolvedValue({ data: [] });
    renderHook(() => useCampaignList({ status: 'open', page: 2, page_size: 50 }), { wrapper });
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith('/campaigns', {
      page: 2,
      page_size: 50,
      status: 'open',
    });
  });
});

describe('useCampaign', () => {
  beforeEach(() => vi.clearAllMocks());

  it('disabled without id', () => {
    const { result } = renderHook(() => useCampaign(undefined), { wrapper });
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches single campaign and unwraps data', async () => {
    mockGet.mockResolvedValue({ data: { id: 'c1' } });
    const { result } = renderHook(() => useCampaign('c1'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockGet).toHaveBeenCalledWith('/campaigns/c1');
    expect(result.current.data).toEqual({ id: 'c1' });
  });
});

describe('useJoinCampaign', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setOnline(true);
  });

  it('posts join when online', async () => {
    mockPost.mockResolvedValue({ data: { id: 'p1' } });
    const { result } = renderHook(() => useJoinCampaign(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ id: 'c1', quantity: 2 });
    });
    expect(mockPost).toHaveBeenCalledWith('/campaigns/c1/join', { quantity: 2 });
  });

  it('enqueues when offline', async () => {
    setOnline(false);
    const { result } = renderHook(() => useJoinCampaign(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ id: 'c1', quantity: 1 });
    });
    expect(mockEnqueueOffline).toHaveBeenCalled();
    expect(mockPost).not.toHaveBeenCalled();
  });
});

describe('useCancelCampaign', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setOnline(true);
  });

  it('posts cancel when online', async () => {
    mockPost.mockResolvedValue({ data: { id: 'c1' } });
    const { result } = renderHook(() => useCancelCampaign(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync('c1');
    });
    expect(mockPost).toHaveBeenCalledWith('/campaigns/c1/cancel', {});
  });

  it('enqueues when offline', async () => {
    setOnline(false);
    const { result } = renderHook(() => useCancelCampaign(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync('c1');
    });
    expect(mockEnqueueOffline).toHaveBeenCalled();
  });
});

describe('useEvaluateCampaign', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setOnline(true);
  });

  it('posts evaluate when online', async () => {
    mockPost.mockResolvedValue({ data: { id: 'c1' } });
    const { result } = renderHook(() => useEvaluateCampaign(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync('c1');
    });
    expect(mockPost).toHaveBeenCalledWith('/campaigns/c1/evaluate', {});
  });

  it('enqueues when offline', async () => {
    setOnline(false);
    const { result } = renderHook(() => useEvaluateCampaign(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync('c1');
    });
    expect(mockEnqueueOffline).toHaveBeenCalled();
  });
});

describe('useCreateCampaign', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setOnline(true);
  });

  it('posts create when online', async () => {
    mockPost.mockResolvedValue({ data: { id: 'c1' } });
    const { result } = renderHook(() => useCreateCampaign(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({
        item_id: 'i1',
        min_quantity: 5,
        cutoff_time: '2026-05-01T00:00:00Z',
      });
    });
    expect(mockPost).toHaveBeenCalledWith(
      '/campaigns',
      expect.objectContaining({ item_id: 'i1', min_quantity: 5 }),
    );
  });

  it('enqueues when offline', async () => {
    setOnline(false);
    const { result } = renderHook(() => useCreateCampaign(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({
        item_id: 'i1',
        min_quantity: 5,
        cutoff_time: '2026-05-01T00:00:00Z',
      });
    });
    expect(mockEnqueueOffline).toHaveBeenCalled();
    expect(mockPost).not.toHaveBeenCalled();
  });
});
