import type { KPIDashboard } from '@/lib/types';

export type { KPIDashboard };

export interface KPIFilters {
  period: 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'yearly';
  location_id?: string;
  coach_id?: string;
  category?: string;
  date_range_start?: string;
  date_range_end?: string;
}

export interface KPICardData {
  label: string;
  value: number;
  formatted_value: string;
  change_percent: number;
  change_direction: 'up' | 'down' | 'flat';
  period: string;
  icon?: string;
}

export interface DashboardSummary {
  kpis: KPIDashboard;
  cards: KPICardData[];
  recent_orders_count: number;
  active_campaigns_count: number;
  low_stock_items_count: number;
  pending_variances_count: number;
}
