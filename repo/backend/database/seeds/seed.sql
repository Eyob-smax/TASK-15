-- Seed data for FitCommerce Operations & Inventory Suite
-- This file is idempotent: safe to run multiple times via INSERT ... ON CONFLICT DO NOTHING.

-- Reserved non-human system actor used as the audit actor for scheduled/automated operations.
-- status = 'inactive' ensures the auth service rejects any login attempt against this account.
-- The login flow checks for inactive status before any credential verification.
INSERT INTO users (id, email, password_hash, salt, role, status, display_name, failed_login_count)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'system@fitcommerce.internal',
    'INVALID_HASH_NO_LOGIN',
    'INVALID_SALT_NO_LOGIN',
    'administrator',
    'inactive',
    'System',
    0
) ON CONFLICT (id) DO NOTHING;

-- Default retention policies
INSERT INTO retention_policies (entity_type, retention_days, description, updated_at) VALUES
    ('financial_records', 2555, 'Financial transaction records including orders, payments, and refunds', NOW()),
    ('procurement_records', 2555, 'Purchase orders, supplier invoices, and landed cost data', NOW()),
    ('access_logs', 730, 'User login attempts, session records, and access audit trails', NOW())
ON CONFLICT (entity_type) DO NOTHING;

-- Default report definitions
INSERT INTO report_definitions (name, report_type, description, allowed_roles) VALUES
    (
        'member_growth',
        'member_growth',
        'Tracks new member sign-ups and total membership growth over time',
        ARRAY['administrator', 'operations_manager']::user_role[]
    ),
    (
        'churn',
        'churn',
        'Measures member cancellation and attrition rates',
        ARRAY['administrator', 'operations_manager']::user_role[]
    ),
    (
        'renewal_rate',
        'renewal_rate',
        'Tracks membership renewal percentages and trends',
        ARRAY['administrator', 'operations_manager']::user_role[]
    ),
    (
        'engagement',
        'engagement',
        'Measures member engagement through activity and participation metrics',
        ARRAY['administrator', 'operations_manager', 'coach']::user_role[]
    ),
    (
        'class_fill_rate',
        'class_fill_rate',
        'Shows class capacity utilization and fill rates',
        ARRAY['administrator', 'operations_manager', 'coach']::user_role[]
    ),
    (
        'coach_productivity',
        'coach_productivity',
        'Evaluates coach performance and session throughput',
        ARRAY['administrator', 'operations_manager']::user_role[]
    ),
    (
        'inventory_summary',
        'inventory_summary',
        'Overview of current inventory levels, status, and valuation',
        ARRAY['administrator', 'operations_manager']::user_role[]
    ),
    (
        'procurement_summary',
        'procurement_summary',
        'Summary of purchase orders, supplier activity, and spending',
        ARRAY['administrator', 'operations_manager', 'procurement_specialist']::user_role[]
    ),
    (
        'landed_cost_report',
        'landed_cost_report',
        'Detailed landed cost breakdown by item, period, and cost component',
        ARRAY['administrator', 'operations_manager', 'procurement_specialist']::user_role[]
    )
ON CONFLICT (name) DO NOTHING;
