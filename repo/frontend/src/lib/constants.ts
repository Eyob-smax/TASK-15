export const USER_ROLES = ['administrator', 'operations_manager', 'procurement_specialist', 'coach', 'member'] as const;
export const ITEM_CONDITIONS = ['new', 'open_box', 'used'] as const;
export const BILLING_MODELS = ['one_time', 'monthly_rental'] as const;
export const ITEM_STATUSES = ['draft', 'published', 'unpublished'] as const;
export const ORDER_STATUSES = ['created', 'paid', 'cancelled', 'refunded', 'auto_closed'] as const;
export const CAMPAIGN_STATUSES = ['active', 'succeeded', 'failed', 'cancelled'] as const;
export const PO_STATUSES = ['created', 'approved', 'received', 'returned', 'voided'] as const;

export const DEFAULT_REFUNDABLE_DEPOSIT = 50.00;
export const AUTO_CLOSE_TIMEOUT_MINUTES = 30;
export const SESSION_IDLE_TIMEOUT_MINUTES = 30;
export const SESSION_ABSOLUTE_TIMEOUT_HOURS = 12;
export const LOGIN_LOCKOUT_THRESHOLD = 5;
export const LOGIN_LOCKOUT_DURATION_MINUTES = 15;
export const VARIANCE_RESOLUTION_BUSINESS_DAYS = 5;
export const VARIANCE_ESCALATION_AMOUNT = 250.00;
export const VARIANCE_ESCALATION_PERCENT = 0.02;
export const RETENTION_FINANCIAL_YEARS = 7;
export const RETENTION_ACCESS_LOG_YEARS = 2;
export const BIOMETRIC_KEY_ROTATION_DAYS = 90;

// Role permission map — which roles can access which modules
export const ROLE_PERMISSIONS: Record<string, string[]> = {
  administrator: ['dashboard', 'catalog', 'inventory', 'group_buys', 'orders', 'procurement', 'reports', 'admin', 'audit', 'backups', 'biometric', 'users'],
  operations_manager: ['dashboard', 'catalog', 'inventory', 'group_buys', 'orders', 'procurement', 'reports'],
  procurement_specialist: ['dashboard', 'catalog', 'procurement', 'reports'],
  coach: ['dashboard', 'group_buys', 'orders', 'reports'],
  member: ['catalog', 'group_buys', 'orders'],
};

// Display labels
export const ROLE_LABELS: Record<string, string> = {
  administrator: 'Administrator',
  operations_manager: 'Operations Manager',
  procurement_specialist: 'Procurement Specialist',
  coach: 'Coach',
  member: 'Member',
};

export const CONDITION_LABELS: Record<string, string> = {
  new: 'New',
  open_box: 'Open Box',
  used: 'Used',
};

export const BILLING_MODEL_LABELS: Record<string, string> = {
  one_time: 'One-Time Purchase',
  monthly_rental: 'Monthly Rental',
};

export const ORDER_STATUS_LABELS: Record<string, string> = {
  created: 'Created',
  paid: 'Paid',
  cancelled: 'Cancelled',
  refunded: 'Refunded',
  auto_closed: 'Auto-Closed',
};

export const CAMPAIGN_STATUS_LABELS: Record<string, string> = {
  active: 'Active',
  succeeded: 'Succeeded',
  failed: 'Failed',
  cancelled: 'Cancelled',
};

export const PO_STATUS_LABELS: Record<string, string> = {
  created: 'Created',
  approved: 'Approved',
  received: 'Received',
  returned: 'Returned',
  voided: 'Voided',
};
