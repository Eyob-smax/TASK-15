-- +goose Up
-- Migration 00009: normalize retention policies to day-based storage and track updates.

ALTER TABLE retention_policies
    RENAME COLUMN retention_years TO retention_days;

ALTER TABLE retention_policies
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE retention_policies
SET
    retention_days = retention_days * 365,
    updated_at = created_at;

-- +goose Down
UPDATE retention_policies
SET retention_days = GREATEST(1, retention_days / 365);

ALTER TABLE retention_policies
    DROP COLUMN IF EXISTS updated_at;

ALTER TABLE retention_policies
    RENAME COLUMN retention_days TO retention_years;
