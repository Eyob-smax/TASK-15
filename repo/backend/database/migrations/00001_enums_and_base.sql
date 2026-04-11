-- +goose Up
-- Migration 00001: Enum types and base infrastructure tables

-- Enum types
CREATE TYPE user_role AS ENUM ('administrator', 'operations_manager', 'procurement_specialist', 'coach', 'member');
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'locked');
CREATE TYPE item_condition AS ENUM ('new', 'open_box', 'used');
CREATE TYPE billing_model AS ENUM ('one_time', 'monthly_rental');
CREATE TYPE item_status AS ENUM ('draft', 'published', 'unpublished');
CREATE TYPE order_status AS ENUM ('created', 'paid', 'cancelled', 'refunded', 'auto_closed');
CREATE TYPE campaign_status AS ENUM ('active', 'succeeded', 'failed', 'cancelled');
CREATE TYPE po_status AS ENUM ('created', 'approved', 'received', 'returned', 'voided');
CREATE TYPE variance_type AS ENUM ('shortage', 'overage', 'price_difference');
CREATE TYPE variance_status AS ENUM ('open', 'resolved');
CREATE TYPE membership_status AS ENUM ('active', 'expired', 'cancelled', 'suspended');
CREATE TYPE export_format AS ENUM ('csv', 'pdf');
CREATE TYPE export_status AS ENUM ('pending', 'processing', 'completed', 'failed');
CREATE TYPE backup_status AS ENUM ('running', 'completed', 'failed');
CREATE TYPE encryption_key_status AS ENUM ('active', 'rotated', 'revoked');

-- Locations
CREATE TABLE locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    address TEXT NOT NULL DEFAULT '',
    timezone VARCHAR(100) NOT NULL DEFAULT 'UTC',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_locations_is_active ON locations(is_active);

-- Retention policies
CREATE TABLE retention_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(100) NOT NULL UNIQUE,
    retention_years INTEGER NOT NULL CHECK (retention_years > 0),
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Encryption keys (for biometric module)
CREATE TABLE encryption_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_reference VARCHAR(255) NOT NULL,
    purpose VARCHAR(50) NOT NULL DEFAULT 'biometric',
    status encryption_key_status NOT NULL DEFAULT 'active',
    activated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rotated_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_encryption_keys_status ON encryption_keys(status);
CREATE INDEX idx_encryption_keys_purpose ON encryption_keys(purpose);

-- +goose Down
-- Drop tables in reverse order
DROP TABLE IF EXISTS encryption_keys;
DROP TABLE IF EXISTS retention_policies;
DROP TABLE IF EXISTS locations;

-- Drop all enum types
DROP TYPE IF EXISTS encryption_key_status;
DROP TYPE IF EXISTS backup_status;
DROP TYPE IF EXISTS export_status;
DROP TYPE IF EXISTS export_format;
DROP TYPE IF EXISTS membership_status;
DROP TYPE IF EXISTS variance_status;
DROP TYPE IF EXISTS variance_type;
DROP TYPE IF EXISTS po_status;
DROP TYPE IF EXISTS campaign_status;
DROP TYPE IF EXISTS order_status;
DROP TYPE IF EXISTS item_status;
DROP TYPE IF EXISTS billing_model;
DROP TYPE IF EXISTS item_condition;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_role;
