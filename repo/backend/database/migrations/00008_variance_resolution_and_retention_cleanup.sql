-- +goose Up
-- Migration 00008: variance resolution metadata and retention-policy cleanup.

ALTER TYPE variance_status ADD VALUE IF NOT EXISTS 'escalated';

ALTER TABLE variance_records
    ADD COLUMN IF NOT EXISTS resolution_action VARCHAR(20) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS quantity_change INTEGER;

DELETE FROM retention_policies
WHERE entity_type = 'audit_events';

-- +goose Down
ALTER TABLE variance_records
    DROP COLUMN IF EXISTS quantity_change,
    DROP COLUMN IF EXISTS resolution_action;

INSERT INTO retention_policies (id, entity_type, retention_years, description, created_at)
SELECT gen_random_uuid(), 'audit_events', 7, 'System-wide audit events with hash-chain integrity', NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM retention_policies WHERE entity_type = 'audit_events'
);
