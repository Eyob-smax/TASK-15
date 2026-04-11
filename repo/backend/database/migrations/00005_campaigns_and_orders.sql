-- +goose Up
-- Migration 00005: Group-buy campaigns, orders, order timeline, fulfillment groups
-- Table creation order ensures all FK references resolve correctly.

CREATE TABLE group_buy_campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id UUID NOT NULL REFERENCES items(id),
    min_quantity INTEGER NOT NULL CHECK (min_quantity > 0),
    current_committed_qty INTEGER NOT NULL DEFAULT 0 CHECK (current_committed_qty >= 0),
    cutoff_time TIMESTAMPTZ NOT NULL,
    status campaign_status NOT NULL DEFAULT 'active',
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    evaluated_at TIMESTAMPTZ
);
CREATE INDEX idx_group_buy_campaigns_item_id ON group_buy_campaigns(item_id);
CREATE INDEX idx_group_buy_campaigns_status ON group_buy_campaigns(status);
CREATE INDEX idx_group_buy_campaigns_cutoff_time ON group_buy_campaigns(cutoff_time);

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    item_id UUID NOT NULL REFERENCES items(id),
    campaign_id UUID REFERENCES group_buy_campaigns(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(10,2) NOT NULL,
    total_amount NUMERIC(10,2) NOT NULL,
    status order_status NOT NULL DEFAULT 'created',
    settlement_marker TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    auto_close_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paid_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ
);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_item_id ON orders(item_id);
CREATE INDEX idx_orders_campaign_id ON orders(campaign_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_auto_close_at ON orders(auto_close_at);

CREATE TABLE group_buy_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES group_buy_campaigns(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    order_id UUID REFERENCES orders(id),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT group_buy_participants_campaign_user_unique UNIQUE (campaign_id, user_id)
);
CREATE INDEX idx_group_buy_participants_campaign_id ON group_buy_participants(campaign_id);
CREATE INDEX idx_group_buy_participants_user_id ON group_buy_participants(user_id);

CREATE TABLE order_timeline_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    performed_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_order_timeline_entries_order_id ON order_timeline_entries(order_id);
CREATE INDEX idx_order_timeline_entries_created_at ON order_timeline_entries(created_at);

-- fulfillment_groups.supplier_id is a plain UUID column here.
-- The FK constraint to the suppliers table is added in migration 00006_procurement.sql
-- after the suppliers table is created.
CREATE TABLE fulfillment_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id UUID,
    warehouse_bin_id UUID REFERENCES warehouse_bins(id),
    pickup_point VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE fulfillment_group_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fulfillment_group_id UUID NOT NULL REFERENCES fulfillment_groups(id) ON DELETE CASCADE,
    order_id UUID NOT NULL REFERENCES orders(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0)
);
CREATE INDEX idx_fulfillment_group_orders_group_id ON fulfillment_group_orders(fulfillment_group_id);
CREATE INDEX idx_fulfillment_group_orders_order_id ON fulfillment_group_orders(order_id);

-- +goose Down
DROP TABLE IF EXISTS fulfillment_group_orders;
DROP TABLE IF EXISTS fulfillment_groups;
DROP TABLE IF EXISTS order_timeline_entries;
DROP TABLE IF EXISTS group_buy_participants;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS group_buy_campaigns;
