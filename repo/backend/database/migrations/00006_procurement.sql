-- +goose Up
-- Migration 00006: Suppliers, purchase orders, variance tracking, landed cost entries,
-- and the deferred FK from fulfillment_groups to suppliers.

CREATE TABLE suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    contact_name VARCHAR(255) NOT NULL DEFAULT '',
    contact_email VARCHAR(255) NOT NULL DEFAULT '',
    contact_phone VARCHAR(50) NOT NULL DEFAULT '',
    address TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT suppliers_name_unique UNIQUE (name)
);
CREATE INDEX idx_suppliers_is_active ON suppliers(is_active);

-- Now that suppliers exists, add the FK constraint deferred from migration 00005
ALTER TABLE fulfillment_groups
    ADD CONSTRAINT fk_fulfillment_groups_supplier
    FOREIGN KEY (supplier_id) REFERENCES suppliers(id);

CREATE TABLE purchase_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id UUID NOT NULL REFERENCES suppliers(id),
    status po_status NOT NULL DEFAULT 'created',
    total_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    created_by UUID NOT NULL REFERENCES users(id),
    approved_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    received_at TIMESTAMPTZ,
    version INTEGER NOT NULL DEFAULT 1
);
CREATE INDEX idx_purchase_orders_supplier_id ON purchase_orders(supplier_id);
CREATE INDEX idx_purchase_orders_status ON purchase_orders(status);
CREATE INDEX idx_purchase_orders_created_by ON purchase_orders(created_by);

CREATE TABLE purchase_order_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_order_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES items(id),
    ordered_quantity INTEGER NOT NULL CHECK (ordered_quantity > 0),
    ordered_unit_price NUMERIC(10,2) NOT NULL CHECK (ordered_unit_price >= 0),
    received_quantity INTEGER,
    received_unit_price NUMERIC(10,2)
);
CREATE INDEX idx_purchase_order_lines_po_id ON purchase_order_lines(purchase_order_id);
CREATE INDEX idx_purchase_order_lines_item_id ON purchase_order_lines(item_id);

CREATE TABLE variance_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    po_line_id UUID NOT NULL REFERENCES purchase_order_lines(id),
    type variance_type NOT NULL,
    expected_value NUMERIC(12,2) NOT NULL,
    actual_value NUMERIC(12,2) NOT NULL,
    difference_amount NUMERIC(12,2) NOT NULL,
    status variance_status NOT NULL DEFAULT 'open',
    resolution_due_date DATE NOT NULL,
    resolved_at TIMESTAMPTZ,
    resolution_notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_variance_records_po_line_id ON variance_records(po_line_id);
CREATE INDEX idx_variance_records_status ON variance_records(status);
CREATE INDEX idx_variance_records_resolution_due_date ON variance_records(resolution_due_date);

CREATE TABLE landed_cost_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id UUID NOT NULL REFERENCES items(id),
    purchase_order_id UUID NOT NULL REFERENCES purchase_orders(id),
    po_line_id UUID NOT NULL REFERENCES purchase_order_lines(id),
    period VARCHAR(20) NOT NULL,
    cost_component VARCHAR(100) NOT NULL,
    raw_amount NUMERIC(12,2) NOT NULL,
    allocated_amount NUMERIC(12,2) NOT NULL,
    allocation_method VARCHAR(50) NOT NULL DEFAULT 'value_weighted',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_landed_cost_entries_item_id ON landed_cost_entries(item_id);
CREATE INDEX idx_landed_cost_entries_period ON landed_cost_entries(period);
CREATE INDEX idx_landed_cost_entries_po_id ON landed_cost_entries(purchase_order_id);

-- +goose Down
DROP TABLE IF EXISTS landed_cost_entries;
DROP TABLE IF EXISTS variance_records;
DROP TABLE IF EXISTS purchase_order_lines;
DROP TABLE IF EXISTS purchase_orders;

-- Drop the FK constraint added to fulfillment_groups before dropping suppliers
ALTER TABLE fulfillment_groups DROP CONSTRAINT IF EXISTS fk_fulfillment_groups_supplier;

DROP TABLE IF EXISTS suppliers;
