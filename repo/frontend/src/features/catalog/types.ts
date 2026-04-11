import type {
  Item,
  ItemCondition,
  ItemStatus,
  BillingModel,
  BatchEditResponse,
  BatchEditResult,
  AvailabilityWindow,
  BlackoutWindow,
} from '@/lib/types';

export type {
  Item,
  ItemCondition,
  ItemStatus,
  BillingModel,
  BatchEditResponse,
  BatchEditResult,
  AvailabilityWindow,
  BlackoutWindow,
};

export interface ItemFilters {
  search?: string;
  category?: string;
  brand?: string;
  condition?: ItemCondition;
  billing_model?: BillingModel;
  status?: ItemStatus;
  location_id?: string;
  min_price?: number;
  max_price?: number;
  in_stock?: boolean;
  tags?: string[];
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface ItemFormData {
  name: string;
  description: string;
  category: string;
  brand: string;
  sku: string;
  condition: ItemCondition;
  billing_model: BillingModel;
  unit_price: number;
  refundable_deposit: number;
  quantity: number;
  location_id: string;
  availability_windows: {
    start_time: string;
    end_time: string;
  }[];
  blackout_windows: {
    start_time: string;
    end_time: string;
  }[];
}

export interface BatchEditFormData {
  rows: {
    item_id: string;
    field: 'name' | 'description' | 'category' | 'brand' | 'condition' | 'billing_model' | 'status' | 'unit_price' | 'refundable_deposit' | 'quantity' | 'availability_windows';
    new_value?: string;
    availability_windows?: {
      start_time: string;
      end_time: string;
    }[];
  }[];
}
