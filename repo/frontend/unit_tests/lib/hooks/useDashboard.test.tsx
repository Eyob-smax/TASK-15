import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useDashboardKPIs } from '@/lib/hooks/useDashboard';

const mockGet = vi.fn();
vi.mock('@/lib/api-client', () => ({
  apiClient: {
    get: (...args: unknown[]) => mockGet(...args),
  },
}));

function wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

describe('useDashboardKPIs', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls /dashboard/kpis with default monthly period when no filters', async () => {
    mockGet.mockResolvedValue({
      data: {
        member_growth: { value: 10 },
        churn: { value: 2 },
        renewal_rate: { value: 90 },
        engagement: { value: 80 },
        class_fill_rate: { value: 75 },
        coach_productivity: { value: 85 },
      },
    });

    const { result } = renderHook(() => useDashboardKPIs(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockGet).toHaveBeenCalledWith('/dashboard/kpis', { period: 'monthly' });
    expect(result.current.data?.member_growth.value).toBe(10);
  });

  it('passes all provided filters through the query params', async () => {
    mockGet.mockResolvedValue({
      data: {
        member_growth: { value: 1 },
        churn: { value: 0 },
        renewal_rate: { value: 100 },
        engagement: { value: 0 },
        class_fill_rate: { value: 0 },
        coach_productivity: { value: 0 },
      },
    });

    const { result } = renderHook(
      () =>
        useDashboardKPIs({
          period: 'weekly',
          location_id: 'loc-1',
          coach_id: 'coach-1',
          category: 'cardio',
          from: '2026-01-01',
          to: '2026-02-01',
        }),
      { wrapper },
    );
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockGet).toHaveBeenCalledWith('/dashboard/kpis', {
      period: 'weekly',
      location_id: 'loc-1',
      coach_id: 'coach-1',
      category: 'cardio',
      from: '2026-01-01',
      to: '2026-02-01',
    });
  });

  it('fills in defaults for missing KPIMetric fields', async () => {
    mockGet.mockResolvedValue({ data: {} });

    const { result } = renderHook(() => useDashboardKPIs(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.member_growth).toEqual({ value: 0 });
    expect(result.current.data?.churn).toEqual({ value: 0 });
    expect(result.current.data?.renewal_rate).toEqual({ value: 0 });
  });

  it('surfaces query errors', async () => {
    mockGet.mockRejectedValue(new Error('network down'));

    const { result } = renderHook(() => useDashboardKPIs(), { wrapper });
    await waitFor(() => expect(result.current.isError).toBe(true));
    expect((result.current.error as Error).message).toContain('network down');
  });
});
