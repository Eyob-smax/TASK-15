import { describe, it, expect } from 'vitest';
import {
  loginSchema,
  createItemSchema,
  createCampaignSchema,
  joinCampaignSchema,
  createOrderSchema,
  payOrderSchema,
  createSupplierSchema,
  resolveVarianceSchema,
  createUserSchema,
  availabilityWindowSchema,
} from '@/lib/validation';

// ─── loginSchema ───────────────────────────────────────────────────────────────

describe('loginSchema', () => {
  it('accepts valid data', () => {
    const result = loginSchema.safeParse({
      email: 'admin@fitcommerce.io',
      password: 'securePass123!',
    });
    expect(result.success).toBe(true);
  });

  it('rejects missing email', () => {
    const result = loginSchema.safeParse({
      password: 'securePass123!',
    });
    expect(result.success).toBe(false);
  });

  it('rejects invalid email format', () => {
    const result = loginSchema.safeParse({
      email: 'not-an-email',
      password: 'securePass123!',
    });
    expect(result.success).toBe(false);
  });

  it('rejects short password', () => {
    const result = loginSchema.safeParse({
      email: 'admin@fitcommerce.io',
      password: 'short',
    });
    expect(result.success).toBe(false);
  });
});

// ─── createItemSchema ──────────────────────────────────────────────────────────

describe('createItemSchema', () => {
  const validItem = {
    name: 'Adjustable Dumbbell Set',
    category: 'free_weights',
    brand: 'IronGrip',
    condition: 'new' as const,
    billing_model: 'one_time' as const,
    quantity: 10,
  };

  it('accepts valid data', () => {
    const result = createItemSchema.safeParse(validItem);
    expect(result.success).toBe(true);
  });

  it('rejects missing name', () => {
    const result = createItemSchema.safeParse({ ...validItem, name: '' });
    expect(result.success).toBe(false);
  });

  it('rejects negative quantity', () => {
    const result = createItemSchema.safeParse({ ...validItem, quantity: -5 });
    expect(result.success).toBe(false);
  });

  it('applies default deposit of 50', () => {
    const result = createItemSchema.safeParse(validItem);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.refundable_deposit).toBe(50);
    }
  });

  it('allows custom deposit', () => {
    const result = createItemSchema.safeParse({ ...validItem, refundable_deposit: 100 });
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.refundable_deposit).toBe(100);
    }
  });

  it('rejects invalid condition', () => {
    const result = createItemSchema.safeParse({ ...validItem, condition: 'broken' });
    expect(result.success).toBe(false);
  });

  it('rejects invalid billing model', () => {
    const result = createItemSchema.safeParse({ ...validItem, billing_model: 'weekly' });
    expect(result.success).toBe(false);
  });
});

// ─── createCampaignSchema ──────────────────────────────────────────────────────

describe('createCampaignSchema', () => {
  const futureDate = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString();

  const validCampaign = {
    item_id: '550e8400-e29b-41d4-a716-446655440000',
    title: 'Summer Dumbbell Group Buy',
    min_quantity: 10,
    cutoff_time: futureDate,
  };

  it('accepts valid data', () => {
    const result = createCampaignSchema.safeParse(validCampaign);
    expect(result.success).toBe(true);
  });

  it('rejects min_quantity of 0', () => {
    const result = createCampaignSchema.safeParse({ ...validCampaign, min_quantity: 0 });
    expect(result.success).toBe(false);
  });

  it('rejects negative min_quantity', () => {
    const result = createCampaignSchema.safeParse({ ...validCampaign, min_quantity: -5 });
    expect(result.success).toBe(false);
  });

  it('requires cutoff_time', () => {
    const { cutoff_time, ...noCutoff } = validCampaign;
    const result = createCampaignSchema.safeParse(noCutoff);
    expect(result.success).toBe(false);
  });

  it('rejects past cutoff_time', () => {
    const pastDate = new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString();
    const result = createCampaignSchema.safeParse({ ...validCampaign, cutoff_time: pastDate });
    expect(result.success).toBe(false);
  });
});

// ─── joinCampaignSchema ────────────────────────────────────────────────────────

describe('joinCampaignSchema', () => {
  it('accepts valid quantity', () => {
    const result = joinCampaignSchema.safeParse({ quantity: 5 });
    expect(result.success).toBe(true);
  });

  it('rejects quantity of 0', () => {
    const result = joinCampaignSchema.safeParse({ quantity: 0 });
    expect(result.success).toBe(false);
  });

  it('rejects negative quantity', () => {
    const result = joinCampaignSchema.safeParse({ quantity: -1 });
    expect(result.success).toBe(false);
  });
});

// ─── createOrderSchema ─────────────────────────────────────────────────────────

describe('createOrderSchema', () => {
  const validOrder = {
    item_id: '550e8400-e29b-41d4-a716-446655440000',
    quantity: 2,
  };

  it('accepts valid data', () => {
    const result = createOrderSchema.safeParse(validOrder);
    expect(result.success).toBe(true);
  });

  it('rejects quantity of 0', () => {
    const result = createOrderSchema.safeParse({ ...validOrder, quantity: 0 });
    expect(result.success).toBe(false);
  });

  it('rejects negative quantity', () => {
    const result = createOrderSchema.safeParse({ ...validOrder, quantity: -1 });
    expect(result.success).toBe(false);
  });

  it('accepts optional campaign_id', () => {
    const result = createOrderSchema.safeParse({
      ...validOrder,
      campaign_id: '550e8400-e29b-41d4-a716-446655440001',
    });
    expect(result.success).toBe(true);
  });
});

// ─── payOrderSchema ────────────────────────────────────────────────────────────

describe('payOrderSchema', () => {
  it('accepts valid settlement marker', () => {
    const result = payOrderSchema.safeParse({ settlement_marker: 'TXN-2026-001' });
    expect(result.success).toBe(true);
  });

  it('rejects empty settlement marker', () => {
    const result = payOrderSchema.safeParse({ settlement_marker: '' });
    expect(result.success).toBe(false);
  });

  it('rejects missing settlement marker', () => {
    const result = payOrderSchema.safeParse({});
    expect(result.success).toBe(false);
  });
});

// ─── createSupplierSchema ──────────────────────────────────────────────────────

describe('createSupplierSchema', () => {
  it('accepts valid data with name only', () => {
    const result = createSupplierSchema.safeParse({ name: 'Rogue Fitness' });
    expect(result.success).toBe(true);
  });

  it('rejects missing name', () => {
    const result = createSupplierSchema.safeParse({});
    expect(result.success).toBe(false);
  });

  it('rejects empty name', () => {
    const result = createSupplierSchema.safeParse({ name: '' });
    expect(result.success).toBe(false);
  });

  it('validates email format if provided', () => {
    const result = createSupplierSchema.safeParse({
      name: 'Rogue Fitness',
      contact_email: 'not-an-email',
    });
    expect(result.success).toBe(false);
  });

  it('accepts valid email if provided', () => {
    const result = createSupplierSchema.safeParse({
      name: 'Rogue Fitness',
      contact_email: 'orders@rogue.com',
    });
    expect(result.success).toBe(true);
  });

  it('accepts empty string for email', () => {
    const result = createSupplierSchema.safeParse({
      name: 'Rogue Fitness',
      contact_email: '',
    });
    expect(result.success).toBe(true);
  });
});

// ─── resolveVarianceSchema ─────────────────────────────────────────────────────

describe('resolveVarianceSchema', () => {
  it('accepts valid resolution notes', () => {
    const result = resolveVarianceSchema.safeParse({
      action: 'return',
      resolution_notes: 'Supplier confirmed shortage and issued credit note for the difference.',
    });
    expect(result.success).toBe(true);
  });

  it('rejects missing resolution notes', () => {
    const result = resolveVarianceSchema.safeParse({});
    expect(result.success).toBe(false);
  });

  it('rejects notes shorter than 10 characters', () => {
    const result = resolveVarianceSchema.safeParse({
      action: 'return',
      resolution_notes: 'Too short',
    });
    expect(result.success).toBe(false);
  });

  it('accepts adjustment resolutions with quantity change', () => {
    const result = resolveVarianceSchema.safeParse({
      action: 'adjustment',
      resolution_notes: 'Count updated after recount.',
      quantity_change: 2,
    });
    expect(result.success).toBe(true);
  });

  it('rejects adjustment resolutions without quantity change', () => {
    const result = resolveVarianceSchema.safeParse({
      action: 'adjustment',
      resolution_notes: 'Count updated after recount.',
    });
    expect(result.success).toBe(false);
  });
});

// ─── createUserSchema ──────────────────────────────────────────────────────────

describe('createUserSchema', () => {
  const validUser = {
    email: 'newuser@fitcommerce.io',
    password: 'securePass123!',
    role: 'member' as const,
    display_name: 'John Doe',
  };

  it('accepts valid user data', () => {
    const result = createUserSchema.safeParse(validUser);
    expect(result.success).toBe(true);
  });

  it('rejects missing email', () => {
    const { email, ...noEmail } = validUser;
    const result = createUserSchema.safeParse(noEmail);
    expect(result.success).toBe(false);
  });

  it('rejects invalid email', () => {
    const result = createUserSchema.safeParse({ ...validUser, email: 'bademail' });
    expect(result.success).toBe(false);
  });

  it('rejects password shorter than 8 characters', () => {
    const result = createUserSchema.safeParse({ ...validUser, password: 'short' });
    expect(result.success).toBe(false);
  });

  it('validates role enum', () => {
    const result = createUserSchema.safeParse({ ...validUser, role: 'superadmin' });
    expect(result.success).toBe(false);
  });

  it('accepts all valid roles', () => {
    const roles = ['administrator', 'operations_manager', 'procurement_specialist', 'coach', 'member'] as const;
    for (const role of roles) {
      const result = createUserSchema.safeParse({ ...validUser, role });
      expect(result.success).toBe(true);
    }
  });
});

// ─── availabilityWindowSchema ──────────────────────────────────────────────────

describe('availabilityWindowSchema', () => {
  const validWindow = {
    item_id: '550e8400-e29b-41d4-a716-446655440000',
    start_time: '2026-05-01T08:00:00Z',
    end_time: '2026-05-31T18:00:00Z',
  };

  it('accepts valid window', () => {
    const result = availabilityWindowSchema.safeParse(validWindow);
    expect(result.success).toBe(true);
  });

  it('rejects end_time before start_time', () => {
    const result = availabilityWindowSchema.safeParse({
      ...validWindow,
      start_time: '2026-05-31T18:00:00Z',
      end_time: '2026-05-01T08:00:00Z',
    });
    expect(result.success).toBe(false);
  });

  it('rejects equal start and end times', () => {
    const result = availabilityWindowSchema.safeParse({
      ...validWindow,
      start_time: '2026-05-01T08:00:00Z',
      end_time: '2026-05-01T08:00:00Z',
    });
    expect(result.success).toBe(false);
  });

  it('rejects missing start_time', () => {
    const { start_time, ...noStart } = validWindow;
    const result = availabilityWindowSchema.safeParse(noStart);
    expect(result.success).toBe(false);
  });

  it('rejects missing end_time', () => {
    const { end_time, ...noEnd } = validWindow;
    const result = availabilityWindowSchema.safeParse(noEnd);
    expect(result.success).toBe(false);
  });
});
