-- +goose Up
-- +goose StatementBegin

-- =============================================================================
-- 004 · persons.created_by_user_id → NULLable
--
-- Rationale: self-registered users create their own Person record before the
-- User row exists, so at INSERT time there is no creator to reference. The
-- original NOT NULL + FK combined with users.person_id NOT NULL produced a
-- chicken-and-egg deadlock on first registration.
--
-- Semantically, a self-registered Person has no distinct "creator" — NULL is
-- the correct representation. Admin-invited Persons continue to carry the
-- creating admin's user id.
-- =============================================================================

ALTER TABLE persons
    ALTER COLUMN created_by_user_id DROP NOT NULL;

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

-- Restoring NOT NULL requires that every row has a non-null value. Any Persons
-- created via self-registration while this migration was in effect will have
-- NULL creators; callers must backfill before rolling back.
ALTER TABLE persons
    ALTER COLUMN created_by_user_id SET NOT NULL;

-- +goose StatementEnd
