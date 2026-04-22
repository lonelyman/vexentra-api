-- +goose Up
-- +goose StatementBegin

-- =============================================================================
-- 001 · Schema Hardening (A + B + C)
--
-- Purpose: close gaps left by GORM AutoMigrate and raise the schema to
-- enterprise-grade consistency before domain modules grow further.
--
-- Scope:
--   A. Standardize — case-insensitive uniqueness, missing slug unique, timestamp defaults
--   B. Data audit & fix — backfill phantom persons row, repair person_id FK graph
--   C. Extend       — add deleted_at to tables that lacked it, add timestamps to audit-less tables
--
-- Pre-audit confirmed ZERO data conflicts for all group B items on 2026-04-21.
-- =============================================================================


-- -----------------------------------------------------------------------------
-- B-0. Data patch — backfill phantom persons row
--      Users, profiles, skills, experiences, portfolio_items, social_links all
--      reference person_id values that never existed in persons (GORM never
--      enforced the FK). Recreate the missing person from profile data before
--      adding any FK constraint.
-- -----------------------------------------------------------------------------

INSERT INTO persons (id, name, linked_user_id, created_by_user_id, created_at, updated_at)
SELECT
    u.person_id,
    COALESCE(pr.display_name, u.username) AS name,
    u.id                                   AS linked_user_id,
    u.id                                   AS created_by_user_id,
    now(),
    now()
FROM users u
LEFT JOIN profiles pr ON pr.person_id = u.person_id
WHERE NOT EXISTS (SELECT 1 FROM persons p WHERE p.id = u.person_id);

-- Cover any stray person_ids referenced by child tables but not by users
INSERT INTO persons (id, name, created_by_user_id, created_at, updated_at)
SELECT DISTINCT child.person_id, 'Unknown', (SELECT id FROM users ORDER BY created_at NULLS LAST LIMIT 1), now(), now()
FROM (
    SELECT person_id FROM profiles        WHERE person_id IS NOT NULL
    UNION SELECT person_id FROM skills         WHERE person_id IS NOT NULL
    UNION SELECT person_id FROM experiences    WHERE person_id IS NOT NULL
    UNION SELECT person_id FROM social_links   WHERE person_id IS NOT NULL
    UNION SELECT person_id FROM portfolio_items WHERE person_id IS NOT NULL
) child
WHERE NOT EXISTS (SELECT 1 FROM persons p WHERE p.id = child.person_id);


-- -----------------------------------------------------------------------------
-- C-1. Extend — add deleted_at to tables that lacked it
-- -----------------------------------------------------------------------------
ALTER TABLE profiles         ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER TABLE experiences      ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER TABLE skills           ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER TABLE social_links     ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER TABLE social_platforms ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;

CREATE INDEX IF NOT EXISTS idx_profiles_deleted_at         ON profiles (deleted_at);
CREATE INDEX IF NOT EXISTS idx_experiences_deleted_at      ON experiences (deleted_at);
CREATE INDEX IF NOT EXISTS idx_skills_deleted_at           ON skills (deleted_at);
CREATE INDEX IF NOT EXISTS idx_social_links_deleted_at     ON social_links (deleted_at);
CREATE INDEX IF NOT EXISTS idx_social_platforms_deleted_at ON social_platforms (deleted_at);


-- -----------------------------------------------------------------------------
-- C-2. Extend — add timestamps to audit-less tables
-- -----------------------------------------------------------------------------
ALTER TABLE user_auths     ADD COLUMN IF NOT EXISTS created_at timestamp with time zone;
ALTER TABLE user_auths     ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone;
ALTER TABLE user_auths     ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;

ALTER TABLE portfolio_tags ADD COLUMN IF NOT EXISTS created_at timestamp with time zone;
ALTER TABLE portfolio_tags ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone;
ALTER TABLE portfolio_tags ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;

ALTER TABLE portfolio_item_tags ADD COLUMN IF NOT EXISTS created_at timestamp with time zone;

-- Backfill timestamps for existing rows (NULL would block subsequent NOT NULL)
UPDATE user_auths          SET created_at = now(), updated_at = now() WHERE created_at IS NULL;
UPDATE portfolio_tags      SET created_at = now(), updated_at = now() WHERE created_at IS NULL;
UPDATE portfolio_item_tags SET created_at = now() WHERE created_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_user_auths_deleted_at     ON user_auths (deleted_at);
CREATE INDEX IF NOT EXISTS idx_portfolio_tags_deleted_at ON portfolio_tags (deleted_at);


-- -----------------------------------------------------------------------------
-- C-3. Standardize — backfill NULL timestamps + enforce NOT NULL DEFAULT now()
-- -----------------------------------------------------------------------------
DO $$
DECLARE
    tbl text;
    tables text[] := ARRAY[
        'persons', 'users', 'user_auths', 'profiles', 'experiences',
        'skills', 'social_platforms', 'social_links',
        'portfolio_items', 'portfolio_tags', 'portfolio_item_tags'
    ];
BEGIN
    FOREACH tbl IN ARRAY tables LOOP
        EXECUTE format('UPDATE %I SET created_at = now() WHERE created_at IS NULL', tbl);
        EXECUTE format('ALTER TABLE %I ALTER COLUMN created_at SET NOT NULL, ALTER COLUMN created_at SET DEFAULT now()', tbl);
    END LOOP;
END$$;

-- updated_at exists on all tables except portfolio_item_tags
DO $$
DECLARE
    tbl text;
    tables text[] := ARRAY[
        'persons', 'users', 'user_auths', 'profiles', 'experiences',
        'skills', 'social_platforms', 'social_links',
        'portfolio_items', 'portfolio_tags'
    ];
BEGIN
    FOREACH tbl IN ARRAY tables LOOP
        EXECUTE format('UPDATE %I SET updated_at = now() WHERE updated_at IS NULL', tbl);
        EXECUTE format('ALTER TABLE %I ALTER COLUMN updated_at SET NOT NULL, ALTER COLUMN updated_at SET DEFAULT now()', tbl);
    END LOOP;
END$$;


-- -----------------------------------------------------------------------------
-- A-1. Standardize — case-insensitive uniqueness on users
-- -----------------------------------------------------------------------------

-- Drop legacy case-sensitive email index, replace with LOWER()
DROP INDEX IF EXISTS idx_users_email_active;
CREATE UNIQUE INDEX idx_users_email_active
    ON users (LOWER(email))
    WHERE deleted_at IS NULL;

-- users.username had NO unique constraint at all — add case-insensitive one
CREATE UNIQUE INDEX idx_users_username_active
    ON users (LOWER(username))
    WHERE deleted_at IS NULL;


-- -----------------------------------------------------------------------------
-- A-2. Standardize — portfolio_items.slug uniqueness
-- -----------------------------------------------------------------------------
CREATE UNIQUE INDEX idx_portfolio_items_slug_active
    ON portfolio_items (slug)
    WHERE deleted_at IS NULL;


-- -----------------------------------------------------------------------------
-- B-1. Data fix — user_auths uniqueness
--      Legacy unique on provider_id alone would collide across providers.
--      Replace with composite (provider, provider_id) scoped to active rows.
-- -----------------------------------------------------------------------------
DROP INDEX IF EXISTS idx_provider_provider_id;
CREATE UNIQUE INDEX idx_user_auths_provider_provider_id
    ON user_auths (provider, provider_id)
    WHERE deleted_at IS NULL AND provider_id IS NOT NULL;


-- -----------------------------------------------------------------------------
-- B-2. Data fix — social_links FK repair
--      GORM wired the FK to profiles(person_id), meaning a profile row had to
--      exist before any social link. Correct target is persons(id).
-- -----------------------------------------------------------------------------
ALTER TABLE social_links DROP CONSTRAINT IF EXISTS fk_profiles_social_links;
ALTER TABLE social_links
    ADD CONSTRAINT fk_social_links_person
    FOREIGN KEY (person_id) REFERENCES persons(id) ON DELETE RESTRICT;

ALTER TABLE social_links
    ADD CONSTRAINT fk_social_links_platform
    FOREIGN KEY (platform_id) REFERENCES social_platforms(id) ON DELETE RESTRICT;


-- -----------------------------------------------------------------------------
-- B-3. Data fix — add missing person_id FKs across the graph
-- -----------------------------------------------------------------------------
ALTER TABLE users
    ADD CONSTRAINT fk_users_person
    FOREIGN KEY (person_id) REFERENCES persons(id) ON DELETE RESTRICT;

ALTER TABLE profiles
    ADD CONSTRAINT fk_profiles_person
    FOREIGN KEY (person_id) REFERENCES persons(id) ON DELETE CASCADE;

ALTER TABLE experiences
    ADD CONSTRAINT fk_experiences_person
    FOREIGN KEY (person_id) REFERENCES persons(id) ON DELETE CASCADE;

ALTER TABLE skills
    ADD CONSTRAINT fk_skills_person
    FOREIGN KEY (person_id) REFERENCES persons(id) ON DELETE CASCADE;

ALTER TABLE portfolio_items
    ADD CONSTRAINT fk_portfolio_items_person
    FOREIGN KEY (person_id) REFERENCES persons(id) ON DELETE CASCADE;


-- -----------------------------------------------------------------------------
-- B-4. Data fix — persons.created_by_user_id / linked_user_id FKs
-- -----------------------------------------------------------------------------
ALTER TABLE persons
    ADD CONSTRAINT fk_persons_created_by_user
    FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE RESTRICT;

ALTER TABLE persons
    ADD CONSTRAINT fk_persons_linked_user
    FOREIGN KEY (linked_user_id) REFERENCES users(id) ON DELETE SET NULL;


-- -----------------------------------------------------------------------------
-- A-3. Standardize — role whitelist on users
--      (Also applied by 002_project_management, but kept here so hardening
--       stands alone if 002 is ever rolled back.)
-- -----------------------------------------------------------------------------
-- Deliberately NOT added here; left to 002 which also performs the
-- 'user' → 'member' data migration. Order: 002 runs after 001.


-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

-- Reverse B-4
ALTER TABLE persons DROP CONSTRAINT IF EXISTS fk_persons_linked_user;
ALTER TABLE persons DROP CONSTRAINT IF EXISTS fk_persons_created_by_user;

-- Reverse B-3
ALTER TABLE portfolio_items DROP CONSTRAINT IF EXISTS fk_portfolio_items_person;
ALTER TABLE skills          DROP CONSTRAINT IF EXISTS fk_skills_person;
ALTER TABLE experiences     DROP CONSTRAINT IF EXISTS fk_experiences_person;
ALTER TABLE profiles        DROP CONSTRAINT IF EXISTS fk_profiles_person;
ALTER TABLE users           DROP CONSTRAINT IF EXISTS fk_users_person;

-- Reverse B-2
ALTER TABLE social_links DROP CONSTRAINT IF EXISTS fk_social_links_platform;
ALTER TABLE social_links DROP CONSTRAINT IF EXISTS fk_social_links_person;
ALTER TABLE social_links
    ADD CONSTRAINT fk_profiles_social_links
    FOREIGN KEY (person_id) REFERENCES profiles(person_id);

-- Reverse B-1
DROP INDEX IF EXISTS idx_user_auths_provider_provider_id;
CREATE UNIQUE INDEX idx_provider_provider_id ON user_auths (provider_id);

-- Reverse A-2
DROP INDEX IF EXISTS idx_portfolio_items_slug_active;

-- Reverse A-1
DROP INDEX IF EXISTS idx_users_username_active;
DROP INDEX IF EXISTS idx_users_email_active;
CREATE UNIQUE INDEX idx_users_email_active ON users (email) WHERE deleted_at IS NULL;

-- Reverse C-3 (loosen NOT NULL / DEFAULT)
DO $$
DECLARE
    tbl text;
    all_tables text[] := ARRAY[
        'persons', 'users', 'user_auths', 'profiles', 'experiences',
        'skills', 'social_platforms', 'social_links',
        'portfolio_items', 'portfolio_tags'
    ];
BEGIN
    FOREACH tbl IN ARRAY all_tables LOOP
        EXECUTE format('ALTER TABLE %I ALTER COLUMN created_at DROP NOT NULL, ALTER COLUMN created_at DROP DEFAULT', tbl);
        EXECUTE format('ALTER TABLE %I ALTER COLUMN updated_at DROP NOT NULL, ALTER COLUMN updated_at DROP DEFAULT', tbl);
    END LOOP;
    EXECUTE 'ALTER TABLE portfolio_item_tags ALTER COLUMN created_at DROP NOT NULL, ALTER COLUMN created_at DROP DEFAULT';
END$$;

-- Reverse C-2 (drop added timestamps)
ALTER TABLE portfolio_item_tags DROP COLUMN IF EXISTS created_at;
ALTER TABLE portfolio_tags      DROP COLUMN IF EXISTS deleted_at, DROP COLUMN IF EXISTS updated_at, DROP COLUMN IF EXISTS created_at;
ALTER TABLE user_auths          DROP COLUMN IF EXISTS deleted_at, DROP COLUMN IF EXISTS updated_at, DROP COLUMN IF EXISTS created_at;

-- Reverse C-1 (drop deleted_at + indexes)
DROP INDEX IF EXISTS idx_social_platforms_deleted_at;
DROP INDEX IF EXISTS idx_social_links_deleted_at;
DROP INDEX IF EXISTS idx_skills_deleted_at;
DROP INDEX IF EXISTS idx_experiences_deleted_at;
DROP INDEX IF EXISTS idx_profiles_deleted_at;
ALTER TABLE social_platforms DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE social_links     DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE skills           DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE experiences      DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE profiles         DROP COLUMN IF EXISTS deleted_at;

-- B-0 backfilled persons row is left in place (cannot safely determine which rows
-- were the phantom vs legitimate once users/profiles reference them).

-- +goose StatementEnd
