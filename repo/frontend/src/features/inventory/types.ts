import type {
  InventorySnapshot,
  InventoryAdjustment,
  WarehouseBin,
  Item,
} from '@/lib/types';

export type {
  InventorySnapshot,
  InventoryAdjustment,
  WarehouseBin,
  Item,
};

export interface InventoryFilters {
  location_id?: string;
  category?: string;
  low_stock_only?: boolean;
  warehouse_bin_id?: string;
  search?: string;
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface AdjustmentFormData {
  item_id: string;
  quantity_change: number;
  reason: string;
}

export interface InventoryOverview {
  total_items: number;
  total_quantity: number;
  low_stock_count: number;
  out_of_stock_count: number;
  total_value: number;
}
