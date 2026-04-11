-- +goose Up
-- Migration 00010: Replace plaintext CAPTCHA answers with derived hashes

ALTER TABLE captcha_challenges
    ADD COLUMN IF NOT EXISTS answer_hash BYTEA,
    ADD COLUMN IF NOT EXISTS answer_salt BYTEA;

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'captcha_challenges'
          AND column_name = 'answer'
    ) THEN
        WITH salted AS (
            SELECT id, gen_random_bytes(16) AS salt
            FROM captcha_challenges
        )
        UPDATE captcha_challenges AS c
        SET
            answer_salt = salted.salt,
            answer_hash = digest(convert_to(lower(trim(c.answer)), 'UTF8') || decode('3a', 'hex') || salted.salt, 'sha256')
        FROM salted
        WHERE c.id = salted.id;
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE captcha_challenges
    ALTER COLUMN answer_hash SET NOT NULL,
    ALTER COLUMN answer_salt SET NOT NULL;

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'captcha_challenges'
          AND column_name = 'answer'
    ) THEN
        ALTER TABLE captcha_challenges
            DROP COLUMN answer;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
ALTER TABLE captcha_challenges
    ADD COLUMN IF NOT EXISTS answer VARCHAR(50);

UPDATE captcha_challenges
SET answer = '[redacted]';

ALTER TABLE captcha_challenges
    ALTER COLUMN answer SET NOT NULL;

ALTER TABLE captcha_challenges
    DROP COLUMN IF EXISTS answer_hash,
    DROP COLUMN IF EXISTS answer_salt;
