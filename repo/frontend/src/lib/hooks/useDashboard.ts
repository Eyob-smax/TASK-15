import { useQuery } from '@tanstack/react-query';
import { apiClient } from '@/lib/api-client';
import type { KPIDashboard } from '@/lib/types';

export type KPIPeriod = 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'yearly';

export interface DashboardFilters {
  period?: KPIPeriod;
  location_id?: string;
  coach_id?: string;
  category?: string;
  from?: string;
  to?: string;
}

// KPIMetricRaw mirrors the backend KPIMetric DTO shape for deserialization.
interface KPIMetricRaw {
  value: number;
  previous_value?: number;
  change_percent?: number;
  period?: string;
}

interface KPIDashboardRaw {
  member_growth: KPIMetricRaw;
  churn: KPIMetricRaw;
  renewal_rate: KPIMetricRaw;
  engagement: KPIMetricRaw;
  class_fill_rate: KPIMetricRaw;
  coach_productivity: KPIMetricRaw;
}

export function useDashboardKPIs(filters: DashboardFilters = {}) {
  return useQuery({
    queryKey: ['dashboard', 'kpis', filters],
    queryFn: async (): Promise<KPIDashboard> => {
      const params: Record<string, string> = { period: filters.period ?? 'monthly' };
      if (filters.location_id) params.location_id = filters.location_id;
      if (filters.coach_id) params.coach_id = filters.coach_id;
      if (filters.category) params.category = filters.category;
      if (filters.from) params.from = filters.from;
      if (filters.to) params.to = filters.to;
      const response = await apiClient.get<{ data: KPIDashboardRaw }>('/dashboard/kpis', params);
      const raw = response.data;
      return {
        member_growth: raw.member_growth || { value: 0 },
        churn: raw.churn || { value: 0 },
        renewal_rate: raw.renewal_rate || { value: 0 },
        engagement: raw.engagement || { value: 0 },
        class_fill_rate: raw.class_fill_rate || { value: 0 },
        coach_productivity: raw.coach_productivity || { value: 0 },
      };
    },
  });
}
