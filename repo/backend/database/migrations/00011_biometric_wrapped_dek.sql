-- +goose Up
ALTER TABLE encryption_keys ADD COLUMN wrapped_dek BYTEA;

-- +goose Down
ALTER TABLE encryption_keys DROP COLUMN wrapped_dek;
