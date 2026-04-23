-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS force_password_change boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS password_changed_at timestamp with time zone;

-- Backfill historical rows so policy can be enabled later without null noise.
UPDATE users
SET password_changed_at = COALESCE(password_changed_at, created_at)
WHERE deleted_at IS NULL AND password_changed_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
  DROP COLUMN IF EXISTS password_changed_at,
  DROP COLUMN IF EXISTS force_password_change;
-- +goose StatementEnd
