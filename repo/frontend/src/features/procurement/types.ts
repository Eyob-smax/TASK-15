import type {
  Supplier,
  PurchaseOrder,
  PurchaseOrderLine,
  POStatus,
  VarianceRecord,
  VarianceType,
  VarianceStatus,
  LandedCostEntry,
} from '@/lib/types';

export type {
  Supplier,
  PurchaseOrder,
  PurchaseOrderLine,
  POStatus,
  VarianceRecord,
  VarianceType,
  VarianceStatus,
  LandedCostEntry,
};

export interface SupplierFilters {
  search?: string;
  is_active?: boolean;
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface POFormData {
  supplier_id: string;
  expected_delivery_date: string;
  currency: string;
  shipping_cost: number;
  insurance_cost: number;
  customs_duty: number;
  other_costs: number;
  notes: string;
  lines: {
    item_id: string;
    ordered_quantity: number;
    ordered_unit_price: number;
  }[];
}

export interface ReceiveFormData {
  lines: {
    po_line_id: string;
    received_quantity: number;
    received_unit_price: number;
  }[];
}

export interface VarianceResolutionFormData {
  action: 'adjustment' | 'return';
  resolution_notes: string;
  quantity_change?: number;
}

export interface LandedCostFilters {
  purchase_order_id?: string;
  item_id?: string;
  supplier_id?: string;
  date_from?: string;
  date_to?: string;
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface ProcurementOverview {
  open_po_count: number;
  open_po_value: number;
  pending_variances: number;
  overdue_deliveries: number;
  active_suppliers: number;
}
