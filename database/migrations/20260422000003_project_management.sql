-- +goose Up
-- +goose StatementBegin

-- =============================================================================
-- 001 · Project Management Module
--
-- Blueprint decisions (locked):
--   · Single-tenant deployment
--   · Person ≠ User (persons/users already managed by GORM AutoMigrate)
--   · Company role: admin | manager | member (migrated from legacy 'user')
--   · Project lifecycle: draft/planned/bidding/active/on_hold/closed
--     + closure_reason (enum) — terminal state only
--   · Project code format: PREFIX-YYYY-NNNN (global sequence, env-configured prefix)
--   · Membership: flat + is_lead (transferable, creator auto-lead)
--   · Client: optional FK to persons + raw snapshot cols (historical)
--   · Financials: transaction_categories (seeded, user-manageable) + project_transactions
--
-- After applying this migration, disable GORM AutoMigrate in application bootstrap.
-- This file becomes the single source of truth for the project management schema.
-- =============================================================================


-- -----------------------------------------------------------------------------
-- 1. Extensions & Sequences
-- -----------------------------------------------------------------------------
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE SEQUENCE IF NOT EXISTS project_code_seq START 1 INCREMENT 1;


-- -----------------------------------------------------------------------------
-- 2. Enum Types
-- -----------------------------------------------------------------------------
CREATE TYPE project_status AS ENUM (
    'draft',
    'planned',
    'bidding',
    'active',
    'on_hold',
    'closed'
);

CREATE TYPE project_closure_reason AS ENUM (
    'won_completed',
    'lost_to_competitor',
    'client_declined',
    'we_withdrew',
    'client_abandoned',
    'budget_cut',
    'cancelled_internal'
);

CREATE TYPE transaction_type AS ENUM (
    'income',
    'expense'
);


-- -----------------------------------------------------------------------------
-- 3. ALTER existing tables (users.role migration + constraint)
-- -----------------------------------------------------------------------------
UPDATE users SET role = 'member' WHERE role = 'user';

ALTER TABLE users
    ADD CONSTRAINT chk_users_role
    CHECK (role IN ('admin', 'manager', 'member'));


-- -----------------------------------------------------------------------------
-- 4. New tables (parents first)
-- -----------------------------------------------------------------------------

-- 4.1 projects -----------------------------------------------------------------
CREATE TABLE projects (
    id                   UUID PRIMARY KEY,
    project_code         TEXT NOT NULL,

    name                 TEXT NOT NULL,
    description          TEXT,

    status               project_status NOT NULL DEFAULT 'draft',
    closure_reason       project_closure_reason,

    -- Client: FK once promoted, raw snapshot frozen as historical record
    client_person_id     UUID REFERENCES persons(id) ON DELETE RESTRICT,
    client_name_raw      TEXT,
    client_email_raw     TEXT,

    -- Dates
    scheduled_start_at   TIMESTAMPTZ,
    deadline_at          TIMESTAMPTZ,
    activated_at         TIMESTAMPTZ,
    closed_at            TIMESTAMPTZ,

    created_by_user_id   UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at           TIMESTAMPTZ,

    -- Code format guard: UPPER-YYYY-NNNN (prefix = alphabetic UPPER only)
    CONSTRAINT chk_projects_code_format
        CHECK (project_code ~ '^[A-Z]+-[0-9]{4}-[0-9]{4}$'),

    -- Terminal state bijection: status='closed' ⇔ closed_at + closure_reason present
    CONSTRAINT chk_projects_closed_consistency
        CHECK (
            (status = 'closed' AND closed_at IS NOT NULL AND closure_reason IS NOT NULL)
            OR
            (status <> 'closed' AND closed_at IS NULL AND closure_reason IS NULL)
        ),

    -- Client required once past pre-sales (active/on_hold/closed)
    CONSTRAINT chk_projects_client_required_when_active
        CHECK (
            status IN ('draft', 'planned', 'bidding')
            OR client_person_id IS NOT NULL
        )
);

-- Partial unique: one active project_code at a time (allow re-use after soft delete? No — codes are permanent business references).
-- Full unique is safer for code: soft-deleted project still "owns" the code.
CREATE UNIQUE INDEX idx_projects_code ON projects (project_code);

CREATE INDEX idx_projects_status_active  ON projects (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_projects_deadline       ON projects (deadline_at) WHERE deleted_at IS NULL AND status <> 'closed';
CREATE INDEX idx_projects_client_person  ON projects (client_person_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_projects_created_by     ON projects (created_by_user_id);
CREATE INDEX idx_projects_updated_at     ON projects (updated_at DESC) WHERE deleted_at IS NULL;


-- 4.2 project_members ---------------------------------------------------------
CREATE TABLE project_members (
    id                   UUID PRIMARY KEY,
    project_id           UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    person_id            UUID NOT NULL REFERENCES persons(id) ON DELETE RESTRICT,

    is_lead              BOOLEAN NOT NULL DEFAULT false,

    added_by_user_id     UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at           TIMESTAMPTZ
);

-- One active membership per (project, person)
CREATE UNIQUE INDEX idx_project_members_active
    ON project_members (project_id, person_id)
    WHERE deleted_at IS NULL;

-- At most one lead per project
CREATE UNIQUE INDEX idx_project_members_one_lead
    ON project_members (project_id)
    WHERE deleted_at IS NULL AND is_lead = true;

CREATE INDEX idx_project_members_person ON project_members (person_id) WHERE deleted_at IS NULL;


-- 4.3 transaction_categories --------------------------------------------------
CREATE TABLE transaction_categories (
    id            UUID PRIMARY KEY,
    code          TEXT NOT NULL,
    name          TEXT NOT NULL,
    type          transaction_type NOT NULL,
    icon_key      TEXT,
    is_system     BOOLEAN NOT NULL DEFAULT false,
    is_active     BOOLEAN NOT NULL DEFAULT true,
    sort_order    INT NOT NULL DEFAULT 0,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ,

    CONSTRAINT chk_tx_categories_code_lower
        CHECK (code = lower(code) AND code ~ '^[a-z0-9_]+$')
);

CREATE UNIQUE INDEX idx_tx_categories_code_active
    ON transaction_categories (code)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_tx_categories_type_sort
    ON transaction_categories (type, sort_order)
    WHERE deleted_at IS NULL AND is_active = true;


-- 4.4 project_transactions ----------------------------------------------------
CREATE TABLE project_transactions (
    id                   UUID PRIMARY KEY,
    project_id           UUID NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    category_id          UUID NOT NULL REFERENCES transaction_categories(id) ON DELETE RESTRICT,

    amount               NUMERIC(15, 2) NOT NULL,
    currency_code        CHAR(3) NOT NULL DEFAULT 'THB',
    note                 TEXT,
    occurred_at          TIMESTAMPTZ NOT NULL,

    created_by_user_id   UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at           TIMESTAMPTZ,

    CONSTRAINT chk_project_tx_amount_nonneg   CHECK (amount >= 0),
    CONSTRAINT chk_project_tx_currency_format CHECK (currency_code ~ '^[A-Z]{3}$')
);

CREATE INDEX idx_project_tx_project_occurred
    ON project_transactions (project_id, occurred_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_project_tx_category
    ON project_transactions (category_id)
    WHERE deleted_at IS NULL;


-- -----------------------------------------------------------------------------
-- 5. Seed data — system transaction categories
--    (is_system=true means user cannot delete; is_active toggleable.)
-- -----------------------------------------------------------------------------
INSERT INTO transaction_categories (id, code, name, type, icon_key, is_system, sort_order) VALUES
    (gen_random_uuid(), 'service_fee',      'ค่าบริการ',             'income',  'briefcase',    true, 10),
    (gen_random_uuid(), 'product_sale',     'ขายสินค้า',             'income',  'package',      true, 20),
    (gen_random_uuid(), 'deposit',          'เงินมัดจำ',             'income',  'wallet',       true, 30),
    (gen_random_uuid(), 'other_income',     'รายรับอื่น ๆ',           'income',  'plus-circle',  true, 99),

    (gen_random_uuid(), 'travel',           'ค่าเดินทาง',             'expense', 'car',          true, 10),
    (gen_random_uuid(), 'stamp_duty',       'ค่าอากรแสตมป์',          'expense', 'stamp',        true, 20),
    (gen_random_uuid(), 'bid_document',     'ค่าซื้อซองประมูล',        'expense', 'file-text',    true, 30),
    (gen_random_uuid(), 'materials',        'ค่าวัสดุอุปกรณ์',         'expense', 'box',          true, 40),
    (gen_random_uuid(), 'subcontract',      'ค่าจ้างผู้รับเหมาช่วง',   'expense', 'users',        true, 50),
    (gen_random_uuid(), 'food',             'ค่าอาหาร',               'expense', 'utensils',     true, 60),
    (gen_random_uuid(), 'certification',    'ค่ารับรอง/เอกสาร',        'expense', 'award',        true, 70),
    (gen_random_uuid(), 'other_expense',    'ค่าใช้จ่ายอื่น ๆ',        'expense', 'minus-circle', true, 99);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS project_transactions;
DROP TABLE IF EXISTS transaction_categories;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;

ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_role;
UPDATE users SET role = 'user' WHERE role = 'member';

DROP TYPE IF EXISTS transaction_type;
DROP TYPE IF EXISTS project_closure_reason;
DROP TYPE IF EXISTS project_status;

DROP SEQUENCE IF EXISTS project_code_seq;

-- (pgcrypto extension left in place — may be used elsewhere.)

-- +goose StatementEnd
