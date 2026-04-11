-- +goose Up
-- Migration 00004: Members and coaches

CREATE TABLE members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) UNIQUE,
    location_id UUID NOT NULL REFERENCES locations(id),
    membership_status membership_status NOT NULL DEFAULT 'active',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    renewal_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_members_user_id ON members(user_id);
CREATE INDEX idx_members_location_id ON members(location_id);
CREATE INDEX idx_members_membership_status ON members(membership_status);

CREATE TABLE coaches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) UNIQUE,
    location_id UUID NOT NULL REFERENCES locations(id),
    specialization VARCHAR(255) NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_coaches_user_id ON coaches(user_id);
CREATE INDEX idx_coaches_location_id ON coaches(location_id);
CREATE INDEX idx_coaches_is_active ON coaches(is_active);

-- +goose Down
DROP TABLE IF EXISTS coaches;
DROP TABLE IF EXISTS members;
