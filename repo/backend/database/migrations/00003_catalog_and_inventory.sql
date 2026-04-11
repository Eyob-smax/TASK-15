-- +goose Up
-- Migration 00003: Catalog items, availability/blackout windows, warehouse bins, inventory tracking, and batch edits

CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category VARCHAR(100) NOT NULL,
    brand VARCHAR(100) NOT NULL DEFAULT '',
    sku VARCHAR(100) NOT NULL DEFAULT '',
    condition item_condition NOT NULL,
    billing_model billing_model NOT NULL,
    unit_price NUMERIC(10,2) NOT NULL DEFAULT 0.00 CHECK (unit_price >= 0),
    refundable_deposit NUMERIC(10,2) NOT NULL DEFAULT 50.00 CHECK (refundable_deposit >= 0),
    quantity INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    status item_status NOT NULL DEFAULT 'draft',
    location_id UUID REFERENCES locations(id),
    created_by UUID NOT NULL REFERENCES users(id),
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_items_status ON items(status);
CREATE INDEX idx_items_category ON items(category);
CREATE INDEX idx_items_brand ON items(brand);
CREATE INDEX idx_items_condition ON items(condition);
CREATE INDEX idx_items_billing_model ON items(billing_model);
CREATE INDEX idx_items_created_by ON items(created_by);
CREATE INDEX idx_items_location_id ON items(location_id);

CREATE TABLE item_availability_windows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    CHECK (end_time > start_time)
);
CREATE INDEX idx_item_availability_windows_item_id ON item_availability_windows(item_id);

CREATE TABLE item_blackout_windows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    CHECK (end_time > start_time)
);
CREATE INDEX idx_item_blackout_windows_item_id ON item_blackout_windows(item_id);

CREATE TABLE warehouse_bins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id UUID NOT NULL REFERENCES locations(id),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT warehouse_bins_location_name_unique UNIQUE (location_id, name)
);
CREATE INDEX idx_warehouse_bins_location_id ON warehouse_bins(location_id);

CREATE TABLE inventory_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id UUID NOT NULL REFERENCES items(id),
    quantity INTEGER NOT NULL,
    location_id UUID REFERENCES locations(id),
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_inventory_snapshots_item_id ON inventory_snapshots(item_id);
CREATE INDEX idx_inventory_snapshots_recorded_at ON inventory_snapshots(recorded_at);

CREATE TABLE inventory_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id UUID NOT NULL REFERENCES items(id),
    quantity_change INTEGER NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_inventory_adjustments_item_id ON inventory_adjustments(item_id);
CREATE INDEX idx_inventory_adjustments_created_at ON inventory_adjustments(created_at);

CREATE TABLE batch_edit_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    total_rows INTEGER NOT NULL DEFAULT 0,
    success_count INTEGER NOT NULL DEFAULT 0,
    failure_count INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE batch_edit_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id UUID NOT NULL REFERENCES batch_edit_jobs(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES items(id),
    field VARCHAR(100) NOT NULL,
    old_value TEXT NOT NULL DEFAULT '',
    new_value TEXT NOT NULL DEFAULT '',
    success BOOLEAN NOT NULL DEFAULT false,
    failure_reason TEXT NOT NULL DEFAULT ''
);
CREATE INDEX idx_batch_edit_results_batch_id ON batch_edit_results(batch_id);

-- +goose Down
DROP TABLE IF EXISTS batch_edit_results;
DROP TABLE IF EXISTS batch_edit_jobs;
DROP TABLE IF EXISTS inventory_adjustments;
DROP TABLE IF EXISTS inventory_snapshots;
DROP TABLE IF EXISTS warehouse_bins;
DROP TABLE IF EXISTS item_blackout_windows;
DROP TABLE IF EXISTS item_availability_windows;
DROP TABLE IF EXISTS items;
