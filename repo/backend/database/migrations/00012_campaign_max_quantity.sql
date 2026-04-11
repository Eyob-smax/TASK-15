-- +goose Up
-- Add optional class capacity field to group-buy campaigns.
-- When set, max_quantity represents the maximum number of participants (class capacity).
-- The class fill rate KPI uses this to compute capacity occupancy.
ALTER TABLE group_buy_campaigns
    ADD COLUMN max_quantity INTEGER CHECK (max_quantity IS NULL OR max_quantity > 0);

-- +goose Down
ALTER TABLE group_buy_campaigns DROP COLUMN max_quantity;
