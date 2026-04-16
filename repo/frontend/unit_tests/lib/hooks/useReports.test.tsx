import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useReportList, useReport, useRunExport, useExport } from '@/lib/hooks/useReports';

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

describe('useReportList', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches /reports and returns result', async () => {
    mockGet.mockResolvedValue({ data: [{ id: 'r1', name: 'Revenue' }] });

    const { result } = renderHook(() => useReportList(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockGet).toHaveBeenCalledWith('/reports');
  });
});

describe('useReport', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('is disabled when id is undefined', () => {
    const { result } = renderHook(() => useReport(undefined), { wrapper });
    expect(result.current.fetchStatus).toBe('idle');
    expect(mockGet).not.toHaveBeenCalled();
  });

  it('fetches by id when enabled', async () => {
    mockGet.mockResolvedValue({ data: { some: 'value' } });

    const { result } = renderHook(() => useReport('r1'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockGet).toHaveBeenCalledWith('/reports/r1/data');
    expect(result.current.data).toEqual({ some: 'value' });
  });
});

describe('useRunExport', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    Object.defineProperty(window.navigator, 'onLine', {
      configurable: true,
      value: true,
    });
  });

  it('posts when online and returns response data', async () => {
    mockPost.mockResolvedValue({ data: { id: 'x1', report_id: 'r1', format: 'csv' } });

    const { result } = renderHook(() => useRunExport(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ report_id: 'r1', format: 'csv' });
    });

    expect(mockPost).toHaveBeenCalledWith('/exports', { report_id: 'r1', format: 'csv' });
  });

  it('enqueues offline mutation when offline', async () => {
    Object.defineProperty(window.navigator, 'onLine', {
      configurable: true,
      value: false,
    });

    const { result } = renderHook(() => useRunExport(), { wrapper });
    let returned: unknown;
    await act(async () => {
      returned = await result.current.mutateAsync({ report_id: 'r2', format: 'pdf' });
    });

    expect(mockEnqueueOffline).toHaveBeenCalled();
    expect(mockPost).not.toHaveBeenCalled();
    expect(returned).toMatchObject({ report_id: 'r2', format: 'pdf', status: 'pending' });
  });
});

describe('useExport', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('is disabled when id is undefined', () => {
    const { result } = renderHook(() => useExport(undefined), { wrapper });
    expect(result.current.fetchStatus).toBe('idle');
    expect(mockGet).not.toHaveBeenCalled();
  });

  it('fetches by id', async () => {
    mockGet.mockResolvedValue({ data: { id: 'x1', status: 'completed' } });

    const { result } = renderHook(() => useExport('x1'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockGet).toHaveBeenCalledWith('/exports/x1');
  });
});
