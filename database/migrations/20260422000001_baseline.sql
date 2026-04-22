-- +goose Up
-- +goose NO TRANSACTION
-- +goose StatementBegin

-- =============================================================================
-- 000 · Baseline Schema (captured from GORM AutoMigrate state)
--
-- Purpose: freeze the current production schema as-is, so future migrations
-- evolve from a known starting point. On existing databases, mark this as
-- applied WITHOUT running it. On fresh databases, goose runs it normally.
--
-- Source: pg_dump -s on vexentra_db (postgres 18.3)
-- Cleaned: removed \restrict markers, SET clutter, search_path fiddling.
-- Added:  pgcrypto extension (needed by 001+).
-- Policy: DO NOT change structure here — hardening goes in 001_schema_hardening.
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";


-- -----------------------------------------------------------------------------
-- Tables (ordered: no-FK parents → children)
-- -----------------------------------------------------------------------------

-- persons -- identity record; created before users
CREATE TABLE IF NOT EXISTS persons (
    id                       uuid NOT NULL,
    name                     character varying(200) NOT NULL,
    invite_email             character varying(254),
    invite_token             character varying(128),
    invite_token_expires_at  timestamp with time zone,
    linked_user_id           uuid,
    created_by_user_id       uuid NOT NULL,
    created_at               timestamp with time zone,
    updated_at               timestamp with time zone,
    deleted_at               timestamp with time zone,
    CONSTRAINT persons_pkey PRIMARY KEY (id)
);

-- users -- auth account linked to a person
CREATE TABLE IF NOT EXISTS users (
    id                                      uuid NOT NULL,
    person_id                               uuid NOT NULL,
    username                                character varying(50) NOT NULL,
    email                                   character varying(254) NOT NULL,
    status                                  character varying(30) DEFAULT 'pending_verification'::character varying NOT NULL,
    last_login_at                           timestamp with time zone,
    is_email_verified                       boolean DEFAULT false NOT NULL,
    email_verification_token                character varying(255),
    email_verification_token_expires_at     timestamp with time zone,
    password_reset_token                    character varying(255),
    password_reset_token_expires_at         timestamp with time zone,
    created_at                              timestamp with time zone,
    updated_at                              timestamp with time zone,
    deleted_at                              timestamp with time zone,
    role                                    character varying(20) DEFAULT 'user'::character varying NOT NULL,
    CONSTRAINT users_pkey PRIMARY KEY (id)
);

-- user_auths -- credential records per (user, provider)
CREATE TABLE IF NOT EXISTS user_auths (
    id              uuid NOT NULL,
    user_id         uuid NOT NULL,
    provider        character varying(30) NOT NULL,
    provider_id     character varying(254),
    secret          character varying(60),
    refresh_token   character varying(512),
    CONSTRAINT user_auths_pkey PRIMARY KEY (id),
    CONSTRAINT fk_users_auths FOREIGN KEY (user_id) REFERENCES users(id)
);

-- profiles -- public profile attached to a person (1:1)
CREATE TABLE IF NOT EXISTS profiles (
    id              uuid NOT NULL,
    person_id       uuid NOT NULL,
    display_name    character varying(100),
    headline        text,
    bio             text,
    location        text,
    avatar_url      text,
    created_at      timestamp with time zone,
    updated_at      timestamp with time zone,
    CONSTRAINT profiles_pkey PRIMARY KEY (id)
);

-- experiences -- work history entries for a person
CREATE TABLE IF NOT EXISTS experiences (
    id              uuid NOT NULL,
    person_id       uuid NOT NULL,
    company         text NOT NULL,
    "position"      text NOT NULL,
    location        character varying(100),
    description     text,
    started_at      timestamp with time zone,
    ended_at        timestamp with time zone,
    is_current      boolean DEFAULT false,
    sort_order      bigint DEFAULT 0,
    created_at      timestamp with time zone,
    updated_at      timestamp with time zone,
    CONSTRAINT experiences_pkey PRIMARY KEY (id)
);

-- skills -- skill entries per person
CREATE TABLE IF NOT EXISTS skills (
    id              uuid NOT NULL,
    person_id       uuid NOT NULL,
    name            text NOT NULL,
    category        text DEFAULT 'other'::text,
    proficiency     bigint DEFAULT 1,
    sort_order      bigint DEFAULT 0,
    created_at      timestamp with time zone,
    updated_at      timestamp with time zone,
    CONSTRAINT skills_pkey PRIMARY KEY (id)
);

-- social_platforms -- catalog of social networks (github, facebook, ...)
CREATE TABLE IF NOT EXISTS social_platforms (
    id              uuid NOT NULL,
    key             text NOT NULL,
    name            text NOT NULL,
    icon_url        text,
    sort_order      bigint DEFAULT 0,
    is_active       boolean DEFAULT true,
    created_at      timestamp with time zone,
    updated_at      timestamp with time zone,
    CONSTRAINT social_platforms_pkey PRIMARY KEY (id)
);

-- social_links -- person's links on each platform
CREATE TABLE IF NOT EXISTS social_links (
    id              uuid NOT NULL,
    person_id       uuid NOT NULL,
    platform_id     uuid NOT NULL,
    url             character varying(512) NOT NULL,
    sort_order      bigint DEFAULT 0,
    created_at      timestamp with time zone,
    updated_at      timestamp with time zone,
    CONSTRAINT social_links_pkey PRIMARY KEY (id),
    -- NOTE: GORM wired this FK to profiles(person_id), not persons(id) — preserved as-is.
    CONSTRAINT fk_profiles_social_links FOREIGN KEY (person_id) REFERENCES profiles(person_id)
);

-- portfolio_tags -- tag catalog
CREATE TABLE IF NOT EXISTS portfolio_tags (
    id      uuid NOT NULL,
    name    text NOT NULL,
    slug    text NOT NULL,
    CONSTRAINT portfolio_tags_pkey PRIMARY KEY (id)
);

-- portfolio_items -- showcase entries
CREATE TABLE IF NOT EXISTS portfolio_items (
    id                  uuid NOT NULL,
    person_id           uuid NOT NULL,
    title               text NOT NULL,
    slug                text NOT NULL,
    summary             text,
    description         text,
    content_markdown    text,
    cover_image_url     text,
    demo_url            text,
    source_url          text,
    status              text DEFAULT 'draft'::text,
    featured            boolean DEFAULT false,
    sort_order          bigint DEFAULT 0,
    started_at          timestamp with time zone,
    ended_at            timestamp with time zone,
    created_at          timestamp with time zone,
    updated_at          timestamp with time zone,
    deleted_at          timestamp with time zone,
    CONSTRAINT portfolio_items_pkey PRIMARY KEY (id)
);

-- portfolio_item_tags -- M:N junction
CREATE TABLE IF NOT EXISTS portfolio_item_tags (
    portfolio_item_id   uuid NOT NULL,
    tag_id              uuid NOT NULL,
    CONSTRAINT portfolio_item_tags_pkey PRIMARY KEY (portfolio_item_id, tag_id),
    CONSTRAINT fk_portfolio_item_tags_portfolio_item_model FOREIGN KEY (portfolio_item_id) REFERENCES portfolio_items(id),
    CONSTRAINT fk_portfolio_item_tags_portfolio_tag_model  FOREIGN KEY (tag_id) REFERENCES portfolio_tags(id)
);


-- -----------------------------------------------------------------------------
-- Indexes (grouped by table)
-- -----------------------------------------------------------------------------

-- persons
CREATE INDEX        IF NOT EXISTS idx_persons_created_by_user_id      ON persons (created_by_user_id);
CREATE INDEX        IF NOT EXISTS idx_persons_deleted_at              ON persons (deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_persons_invite_email_active     ON persons (invite_email)  WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_persons_invite_token            ON persons (invite_token)  WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_persons_linked_user             ON persons (linked_user_id) WHERE deleted_at IS NULL;

-- users
CREATE INDEX        IF NOT EXISTS idx_users_deleted_at                            ON users (deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_active                          ON users (email)                    WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_verification_token_active       ON users (email_verification_token) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_password_reset_token_active           ON users (password_reset_token)     WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_person                                ON users (person_id)                WHERE deleted_at IS NULL;

-- user_auths
CREATE INDEX        IF NOT EXISTS idx_user_auths_user_id       ON user_auths (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_provider_provider_id     ON user_auths (provider_id);

-- profiles
CREATE UNIQUE INDEX IF NOT EXISTS idx_profiles_person_id       ON profiles (person_id);

-- experiences
CREATE INDEX        IF NOT EXISTS idx_experiences_person_id    ON experiences (person_id);

-- skills
CREATE INDEX        IF NOT EXISTS idx_skills_person_id         ON skills (person_id);

-- social_platforms
CREATE UNIQUE INDEX IF NOT EXISTS idx_social_platforms_key     ON social_platforms (key);

-- social_links
CREATE UNIQUE INDEX IF NOT EXISTS idx_social_links_person_platform  ON social_links (person_id, platform_id);

-- portfolio_tags
CREATE UNIQUE INDEX IF NOT EXISTS idx_portfolio_tags_name      ON portfolio_tags (name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_portfolio_tags_slug      ON portfolio_tags (slug);

-- portfolio_items
CREATE INDEX        IF NOT EXISTS idx_portfolio_items_deleted_at ON portfolio_items (deleted_at);
CREATE INDEX        IF NOT EXISTS idx_portfolio_items_person_id  ON portfolio_items (person_id);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS portfolio_item_tags;
DROP TABLE IF EXISTS portfolio_items;
DROP TABLE IF EXISTS portfolio_tags;
DROP TABLE IF EXISTS social_links;
DROP TABLE IF EXISTS social_platforms;
DROP TABLE IF EXISTS skills;
DROP TABLE IF EXISTS experiences;
DROP TABLE IF EXISTS profiles;
DROP TABLE IF EXISTS user_auths;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS persons;

-- pgcrypto extension left in place.

-- +goose StatementEnd
