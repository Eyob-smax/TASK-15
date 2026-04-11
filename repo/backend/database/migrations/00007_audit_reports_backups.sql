-- +goose Up
-- Migration 00007: Audit events (hash-chained), report definitions, export jobs, and backup runs

CREATE TABLE audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID NOT NULL,
    actor_id UUID NOT NULL REFERENCES users(id),
    details JSONB NOT NULL DEFAULT '{}',
    -- integrity_hash and previous_hash form a hash chain for tamper evidence
    integrity_hash VARCHAR(64) NOT NULL,
    previous_hash VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_audit_events_event_type ON audit_events(event_type);
CREATE INDEX idx_audit_events_entity_type ON audit_events(entity_type);
CREATE INDEX idx_audit_events_entity_id ON audit_events(entity_id);
CREATE INDEX idx_audit_events_actor_id ON audit_events(actor_id);
CREATE INDEX idx_audit_events_created_at ON audit_events(created_at);

CREATE TABLE report_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    report_type VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    allowed_roles user_role[] NOT NULL,
    filters JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT report_definitions_name_unique UNIQUE (name)
);

CREATE TABLE export_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES report_definitions(id),
    format export_format NOT NULL,
    filename VARCHAR(255) NOT NULL,
    status export_status NOT NULL DEFAULT 'pending',
    file_path TEXT NOT NULL DEFAULT '',
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);
CREATE INDEX idx_export_jobs_report_id ON export_jobs(report_id);
CREATE INDEX idx_export_jobs_status ON export_jobs(status);
CREATE INDEX idx_export_jobs_created_by ON export_jobs(created_by);

CREATE TABLE backup_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    archive_path TEXT NOT NULL,
    checksum VARCHAR(128) NOT NULL,
    checksum_algorithm VARCHAR(20) NOT NULL DEFAULT 'sha256',
    encryption_key_ref VARCHAR(255) NOT NULL,
    status backup_status NOT NULL,
    file_size BIGINT NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ
);
CREATE INDEX idx_backup_runs_status ON backup_runs(status);
CREATE INDEX idx_backup_runs_started_at ON backup_runs(started_at);

-- +goose Down
DROP TABLE IF EXISTS backup_runs;
DROP TABLE IF EXISTS export_jobs;
DROP TABLE IF EXISTS report_definitions;
DROP TABLE IF EXISTS audit_events;
