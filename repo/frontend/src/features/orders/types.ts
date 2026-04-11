import type {
  Order,
  OrderStatus,
  OrderTimelineEntry,
  FulfillmentGroup,
} from '@/lib/types';

export type {
  Order,
  OrderStatus,
  OrderTimelineEntry,
  FulfillmentGroup,
};

export interface OrderFilters {
  status?: OrderStatus;
  user_id?: string;
  item_id?: string;
  campaign_id?: string;
  search?: string;
  created_after?: string;
  created_before?: string;
  min_total?: number;
  max_total?: number;
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface OrderNoteFormData {
  note: string;
}

export interface OrderSummary {
  total_orders: number;
  total_revenue: number;
  pending_payment_count: number;
  cancelled_count: number;
  refunded_amount: number;
}
