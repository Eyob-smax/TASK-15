// ─── Enums ──────────────────────────────────────────────────────────────────

export type UserRole = 'administrator' | 'operations_manager' | 'procurement_specialist' | 'coach' | 'member';
export type UserStatus = 'active' | 'inactive' | 'locked';
export type ItemCondition = 'new' | 'open_box' | 'used';
export type BillingModel = 'one_time' | 'monthly_rental';
export type ItemStatus = 'draft' | 'published' | 'unpublished';
export type OrderStatus = 'created' | 'paid' | 'cancelled' | 'refunded' | 'auto_closed';
export type CampaignStatus = 'active' | 'succeeded' | 'failed' | 'cancelled';
export type POStatus = 'created' | 'approved' | 'received' | 'returned' | 'voided';
export type VarianceType = 'shortage' | 'overage' | 'price_difference';
export type VarianceStatus = 'open' | 'resolved' | 'escalated';
export type MembershipStatus = 'active' | 'expired' | 'cancelled' | 'suspended';
export type ExportFormat = 'csv' | 'pdf';
export type ExportStatus = 'pending' | 'processing' | 'completed' | 'failed';
export type BackupStatus = 'running' | 'completed' | 'failed';

// ─── Entities ───────────────────────────────────────────────────────────────

export interface User {
  id: string;
  email: string;
  display_name: string;
  role: UserRole;
  status: UserStatus;
  location_id: string | null;
  failed_login_attempts: number;
  locked_until: string | null;
  last_login_at: string | null;
  password_changed_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface Session {
  idle_expires_at: string;
  absolute_expires_at: string;
}

export interface Location {
  id: string;
  name: string;
  address: string;
  timezone: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Member {
  id: string;
  user_id: string;
  location_id: string;
  membership_status: MembershipStatus;
  joined_at: string;
  renewal_date: string | null;
  created_at: string;
  updated_at: string;
}

export interface Coach {
  id: string;
  user_id: string;
  location_id: string;
  specialization: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Item {
  id: string;
  sku: string;
  name: string;
  description: string;
  category: string;
  brand: string;
  condition: ItemCondition;
  unit_price: number;
  refundable_deposit: number;
  billing_model: BillingModel;
  status: ItemStatus;
  quantity: number;
  location_id: string | null;
  created_by: string;
  created_at: string;
  updated_at: string;
  version: number;
  availability_windows?: AvailabilityWindow[];
  blackout_windows?: BlackoutWindow[];
}

export interface AvailabilityWindow {
  id: string;
  start_time: string;
  end_time: string;
}

export interface BlackoutWindow {
  id: string;
  start_time: string;
  end_time: string;
}

export interface InventorySnapshot {
  id: string;
  item_id: string;
  quantity_on_hand: number;
  snapshot_date: string;
  location_id: string | null;
}

export interface InventoryAdjustment {
  id: string;
  item_id: string;
  quantity_change: number;
  reason: string;
  created_by: string;
  created_at: string;
}

export interface WarehouseBin {
  id: string;
  location_id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface BatchEditResult {
  item_id: string;
  field: string;
  old_value: string;
  new_value: string;
  success: boolean;
  failure_reason?: string;
}

export interface BatchEditResponse {
  job_id: string;
  total_rows: number;
  success_count: number;
  failure_count: number;
  results: BatchEditResult[];
}

export interface GroupBuyCampaign {
  id: string;
  item_id: string;
  min_quantity: number;
  current_committed_qty: number;
  cutoff_time: string;
  status: CampaignStatus;
  created_by: string;
  created_at: string;
  evaluated_at: string | null;
}

export interface GroupBuyParticipant {
  id: string;
  campaign_id: string;
  user_id: string;
  quantity: number;
  joined_at: string;
  order_id: string | null;
}

export interface Order {
  id: string;
  user_id: string;
  item_id: string;
  campaign_id: string | null;
  quantity: number;
  unit_price: number;
  total_amount: number;
  status: OrderStatus;
  settlement_marker: string;
  notes: string;
  auto_close_at: string | null;
  paid_at: string | null;
  cancelled_at: string | null;
  refunded_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface OrderTimelineEntry {
  id: string;
  order_id: string;
  action: string;
  performed_by: string;
  description: string;
  created_at: string;
}

export interface FulfillmentGroup {
  id: string;
  campaign_id: string;
  supplier_id: string | null;
  purchase_order_id: string | null;
  status: string;
  shipped_at: string | null;
  delivered_at: string | null;
  tracking_number: string;
  carrier: string;
  created_at: string;
  updated_at: string;
}

export interface Supplier {
  id: string;
  name: string;
  contact_name: string;
  contact_email: string;
  contact_phone: string;
  address: string;
  payment_terms: string;
  lead_time_days: number;
  rating: number;
  is_active: boolean;
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface PurchaseOrder {
  id: string;
  supplier_id: string;
  status: POStatus;
  total_amount: number;
  approved_by: string | null;
  approved_at: string | null;
  received_at: string | null;
  created_by: string;
  created_at: string;
  version: number;
  lines?: PurchaseOrderLine[];
}

export interface PurchaseOrderLine {
  id: string;
  purchase_order_id: string;
  item_id: string;
  ordered_quantity: number;
  ordered_unit_price: number;
  received_quantity: number | null;
  received_unit_price: number | null;
}

export interface VarianceRecord {
  id: string;
  po_line_id: string;
  type: VarianceType;
  expected_value: number;
  actual_value: number;
  difference_amount: number;
  status: VarianceStatus;
  resolution_notes: string;
  resolved_at: string | null;
  resolution_action: 'adjustment' | 'return' | '';
  quantity_change: number | null;
  requires_escalation: boolean;
  is_overdue: boolean;
  created_at: string;
}

export interface LandedCostEntry {
  id: string;
  item_id: string;
  purchase_order_id: string;
  po_line_id: string;
  period: string;
  cost_component: string;
  raw_amount: number;
  allocated_amount: number;
  allocation_method: string;
  created_at: string;
}

export interface AuditEvent {
  id: string;
  event_type: string;
  entity_type: string;
  entity_id: string;
  actor_id: string;
  details: Record<string, unknown>;
  integrity_hash: string;
  created_at: string;
}

export interface BackupRun {
  id: string;
  archive_path: string;
  checksum: string;
  checksum_algorithm: string;
  status: BackupStatus;
  file_size: number;
  started_at: string;
  completed_at: string | null;
}

export interface ReportDefinition {
  id: string;
  name: string;
  description: string;
  report_type: string;
  query_template: string;
  parameters: string;
  allowed_roles: UserRole[];
  is_active: boolean;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface ExportJob {
  id: string;
  report_id: string;
  format: ExportFormat;
  filename: string;
  status: ExportStatus;
  created_by: string;
  created_at: string;
  completed_at: string | null;
}

export interface RetentionPolicy {
  id: string;
  entity_type: string;
  retention_days: number;
  updated_at: string;
}

// ─── API Types ──────────────────────────────────────────────────────────────

export interface ApiErrorResponse {
  error: {
    code: string;
    message: string;
    details?: { field?: string; message: string }[];
  };
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    page: number;
    page_size: number;
    total_count: number;
    total_pages: number;
  };
}

export interface KPIMetric {
  value: number;
  previous_value?: number;
  change_percent?: number;
  period?: string;
}

export interface KPIDashboard {
  member_growth: KPIMetric;
  churn: KPIMetric;
  renewal_rate: KPIMetric;
  engagement: KPIMetric;
  class_fill_rate: KPIMetric;
  coach_productivity: KPIMetric;
}
