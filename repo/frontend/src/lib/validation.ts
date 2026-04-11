import { z } from 'zod';

const defaultedNumber = (defaultValue: number, message: string) =>
  z.preprocess(
    (value) => (value === '' || value == null || Number.isNaN(value) ? defaultValue : value),
    z.number().min(0, message),
  );

const defaultedInteger = (defaultValue: number, message: string) =>
  z.preprocess(
    (value) => (value === '' || value == null || Number.isNaN(value) ? defaultValue : value),
    z.number().int('Quantity must be a whole number').min(0, message),
  );

const itemWindowSchema = z.object({
  start_time: z.string().min(1, 'Start time is required'),
  end_time: z.string().min(1, 'End time is required'),
}).superRefine((window, ctx) => {
  if (new Date(window.end_time) <= new Date(window.start_time)) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path: ['end_time'],
      message: 'End time must be after start time',
    });
  }
});

// ─── Auth ───────────────────────────────────────────────────────────────────

export const loginSchema = z.object({
  email: z.string().min(1, 'Email is required').email('Invalid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
});

// ─── Catalog / Items ────────────────────────────────────────────────────────

export const createItemSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255, 'Name must be 255 characters or less'),
  description: z.string().default(''),
  category: z.string().min(1, 'Category is required'),
  brand: z.string().min(1, 'Brand is required'),
  sku: z.string().default(''),
  condition: z.enum(['new', 'open_box', 'used'], { required_error: 'Condition is required' }),
  billing_model: z.enum(['one_time', 'monthly_rental'], { required_error: 'Billing model is required' }),
  unit_price: defaultedNumber(0, 'Price must be non-negative'),
  refundable_deposit: defaultedNumber(50, 'Deposit must be non-negative'),
  quantity: defaultedInteger(0, 'Quantity must be non-negative'),
  location_id: z.union([z.string().uuid('Invalid location'), z.literal('')]).default(''),
  availability_windows: z.array(itemWindowSchema).default([]),
  blackout_windows: z.array(itemWindowSchema).default([]),
});

export const updateItemSchema = createItemSchema.partial();

export const batchEditRowSchema = z.object({
  itemId: z.string().uuid('Invalid item ID'),
  field: z.enum(['name', 'description', 'category', 'brand', 'condition', 'billing_model', 'status', 'unit_price', 'refundable_deposit', 'quantity', 'availability_windows'], {
    required_error: 'Field is required',
  }),
  newValue: z.string().optional(),
  availabilityWindows: z.array(itemWindowSchema).optional(),
}).superRefine((value, ctx) => {
  if (value.field === 'availability_windows') {
    return;
  }
  if (!value.newValue || !value.newValue.trim()) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path: ['newValue'],
      message: 'New value is required',
    });
  }
});

// ─── Group Buy Campaigns ────────────────────────────────────────────────────

export const createCampaignSchema = z.object({
  item_id: z.string().uuid('Invalid item ID'),
  min_quantity: z.number().int('Must be a whole number').positive('Minimum quantity must be greater than 0'),
  cutoff_time: z.string().refine(
    (val) => new Date(val) > new Date(),
    { message: 'Cutoff time must be in the future' },
  ),
});

export const joinCampaignSchema = z.object({
  quantity: z.number().int('Must be a whole number').positive('Quantity must be greater than 0'),
});

// ─── Orders ─────────────────────────────────────────────────────────────────

export const createOrderSchema = z.object({
  item_id: z.string().uuid('Invalid item ID'),
  quantity: z.number().int('Must be a whole number').positive('Quantity must be greater than 0'),
  campaign_id: z.string().uuid('Invalid campaign ID').optional().nullable(),
});

export const payOrderSchema = z.object({
  settlement_marker: z.string().min(1, 'Settlement marker is required'),
});

// ─── Procurement ────────────────────────────────────────────────────────────

export const createSupplierSchema = z.object({
  name: z.string().min(1, 'Supplier name is required'),
  contact_name: z.string().optional().default(''),
  contact_email: z.string().email('Invalid email').optional().or(z.literal('')),
  contact_phone: z.string().optional().default(''),
  address: z.string().optional().default(''),
  payment_terms: z.string().optional().default(''),
  lead_time_days: z.number().int().min(0).optional(),
  notes: z.string().optional().default(''),
});

const purchaseOrderLineSchema = z.object({
  item_id: z.string().uuid('Invalid item ID'),
  ordered_quantity: z.number().int().positive('Quantity must be greater than 0'),
  ordered_unit_price: z.number().min(0, 'Unit price must be non-negative'),
});

export const createPurchaseOrderSchema = z.object({
  supplier_id: z.string().uuid('Invalid supplier ID'),
  expected_delivery_date: z.string().optional(),
  currency: z.string().default('USD'),
  shipping_cost: z.number().min(0).optional().default(0),
  insurance_cost: z.number().min(0).optional().default(0),
  customs_duty: z.number().min(0).optional().default(0),
  other_costs: z.number().min(0).optional().default(0),
  notes: z.string().optional().default(''),
  lines: z.array(purchaseOrderLineSchema).min(1, 'At least one line item is required'),
});

const receiveLineSchema = z.object({
  po_line_id: z.string().uuid('Invalid line ID'),
  received_quantity: z.number().int().min(0, 'Received quantity must be non-negative'),
  received_unit_price: z.number().min(0, 'Received unit price must be non-negative'),
});

export const receivePurchaseOrderSchema = z.object({
  lines: z.array(receiveLineSchema).min(1, 'At least one line item is required'),
});

export const resolveVarianceSchema = z.object({
  action: z.enum(['adjustment', 'return'], { required_error: 'Resolution action is required' }),
  resolution_notes: z.string().min(10, 'Resolution notes must be at least 10 characters'),
  quantity_change: z.number().int().optional(),
}).superRefine((value, ctx) => {
  if (value.action === 'adjustment' && value.quantity_change == null) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path: ['quantity_change'],
      message: 'Quantity change is required for adjustment resolutions',
    });
  }
  if (value.action === 'return' && value.quantity_change != null) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path: ['quantity_change'],
      message: 'Quantity change is only valid for adjustment resolutions',
    });
  }
});

// ─── Reports & Exports ─────────────────────────────────────────────────────

export const createExportSchema = z.object({
  report_id: z.string().uuid('Invalid report ID'),
  format: z.enum(['csv', 'pdf'], { required_error: 'Format is required' }),
  parameters: z.record(z.string()).optional(),
});

// ─── Admin ──────────────────────────────────────────────────────────────────

export const createUserSchema = z.object({
  email: z.string().min(1, 'Email is required').email('Invalid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  role: z.enum(['administrator', 'operations_manager', 'procurement_specialist', 'coach', 'member'], {
    required_error: 'Role is required',
  }),
  display_name: z.string().min(1, 'Display name is required'),
  location_id: z.string().uuid('Invalid location ID').optional().nullable(),
});

export const createLocationSchema = z.object({
  name: z.string().min(1, 'Location name is required'),
  address: z.string().optional().default(''),
  timezone: z.string().optional().default('UTC'),
});

// ─── Inventory ──────────────────────────────────────────────────────────────

export const inventoryAdjustmentSchema = z.object({
  item_id: z.string().uuid('Invalid item ID'),
  quantity_change: z.number().int('Must be a whole number').refine(
    (val) => val !== 0,
    { message: 'Quantity change must not be zero' },
  ),
  reason: z.string().min(1, 'Reason is required'),
});

export const availabilityWindowSchema = z.object({
  item_id: z.string().uuid('Invalid item ID'),
  start_time: z.string().min(1, 'Start time is required'),
  end_time: z.string().min(1, 'End time is required'),
  recurring: z.boolean().optional().default(false),
  recurrence_rule: z.string().optional().default(''),
}).refine(
  (data) => new Date(data.end_time) > new Date(data.start_time),
  { message: 'End time must be after start time', path: ['end_time'] },
);

export const blackoutWindowSchema = z.object({
  item_id: z.string().uuid('Invalid item ID'),
  start_time: z.string().min(1, 'Start time is required'),
  end_time: z.string().min(1, 'End time is required'),
  reason: z.string().optional().default(''),
}).refine(
  (data) => new Date(data.end_time) > new Date(data.start_time),
  { message: 'End time must be after start time', path: ['end_time'] },
);

// ─── Inferred Types ─────────────────────────────────────────────────────────

export type LoginFormData = z.infer<typeof loginSchema>;
export type CreateItemFormData = z.infer<typeof createItemSchema>;
export type UpdateItemFormData = z.infer<typeof updateItemSchema>;
export type BatchEditRowFormData = z.infer<typeof batchEditRowSchema>;
export type CreateCampaignFormData = z.infer<typeof createCampaignSchema>;
export type JoinCampaignFormData = z.infer<typeof joinCampaignSchema>;
export type CreateOrderFormData = z.infer<typeof createOrderSchema>;
export type PayOrderFormData = z.infer<typeof payOrderSchema>;
export type CreateSupplierFormData = z.infer<typeof createSupplierSchema>;
export type CreatePurchaseOrderFormData = z.infer<typeof createPurchaseOrderSchema>;
export type ReceivePurchaseOrderFormData = z.infer<typeof receivePurchaseOrderSchema>;
export type ResolveVarianceFormData = z.infer<typeof resolveVarianceSchema>;
export type CreateExportFormData = z.infer<typeof createExportSchema>;
export type CreateUserFormData = z.infer<typeof createUserSchema>;
export type CreateLocationFormData = z.infer<typeof createLocationSchema>;
export type InventoryAdjustmentFormData = z.infer<typeof inventoryAdjustmentSchema>;
export type AvailabilityWindowFormData = z.infer<typeof availabilityWindowSchema>;
export type BlackoutWindowFormData = z.infer<typeof blackoutWindowSchema>;
