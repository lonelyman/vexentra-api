-- +goose Up
-- +goose StatementBegin

-- Backup current users table before normalization (dev-safe snapshot).
CREATE TABLE IF NOT EXISTS users_backup_20260424_role_status_refactor AS
SELECT *
FROM users;

-- Master: role
CREATE TABLE IF NOT EXISTS user_role_master (
    code         VARCHAR(20) PRIMARY KEY,
    label_th     TEXT NOT NULL,
    label_en     TEXT NOT NULL,
    sort_order   INT NOT NULL DEFAULT 0,
    is_active    BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_role_master_active_sort
    ON user_role_master (is_active, sort_order);

INSERT INTO user_role_master (code, label_th, label_en, sort_order, is_active)
VALUES
    ('admin', 'ผู้ดูแลระบบ', 'Administrator', 10, true),
    ('manager', 'ผู้จัดการ', 'Manager', 20, true),
    ('member', 'สมาชิก', 'Member', 30, true)
ON CONFLICT (code) DO UPDATE
SET
    label_th = EXCLUDED.label_th,
    label_en = EXCLUDED.label_en,
    sort_order = EXCLUDED.sort_order,
    is_active = EXCLUDED.is_active,
    updated_at = now();

-- Master: status
CREATE TABLE IF NOT EXISTS user_status_master (
    code         VARCHAR(30) PRIMARY KEY,
    label_th     TEXT NOT NULL,
    label_en     TEXT NOT NULL,
    sort_order   INT NOT NULL DEFAULT 0,
    is_active    BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_status_master_active_sort
    ON user_status_master (is_active, sort_order);

INSERT INTO user_status_master (code, label_th, label_en, sort_order, is_active)
VALUES
    ('pending_verification', 'รอยืนยันอีเมล', 'Pending Verification', 10, true),
    ('active', 'ใช้งานได้', 'Active', 20, true),
    ('banned', 'ระงับการใช้งาน', 'Banned', 30, true)
ON CONFLICT (code) DO UPDATE
SET
    label_th = EXCLUDED.label_th,
    label_en = EXCLUDED.label_en,
    sort_order = EXCLUDED.sort_order,
    is_active = EXCLUDED.is_active,
    updated_at = now();

-- Normalize legacy values before adding FK
UPDATE users
SET role = lower(trim(role))
WHERE role IS NOT NULL;

UPDATE users
SET status = lower(trim(status))
WHERE status IS NOT NULL;

UPDATE users
SET role = 'member'
WHERE role = 'user' OR role IS NULL OR role = '' OR role NOT IN ('admin', 'manager', 'member');

UPDATE users
SET status = 'pending_verification'
WHERE status IS NULL OR status = '' OR status NOT IN ('pending_verification', 'active', 'banned');

ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_role;

ALTER TABLE users
    ADD CONSTRAINT fk_users_role_master
    FOREIGN KEY (role) REFERENCES user_role_master(code)
    ON UPDATE RESTRICT
    ON DELETE RESTRICT;

ALTER TABLE users
    ADD CONSTRAINT fk_users_status_master
    FOREIGN KEY (status) REFERENCES user_status_master(code)
    ON UPDATE RESTRICT
    ON DELETE RESTRICT;

ALTER TABLE users ALTER COLUMN role SET DEFAULT 'member';
ALTER TABLE users ALTER COLUMN status SET DEFAULT 'pending_verification';

CREATE INDEX IF NOT EXISTS idx_users_role_active ON users (role) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_status_active ON users (status) WHERE deleted_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_users_status_active;
DROP INDEX IF EXISTS idx_users_role_active;

ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_status_master;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_role_master;

ALTER TABLE users ALTER COLUMN role SET DEFAULT 'member';
ALTER TABLE users ALTER COLUMN status SET DEFAULT 'pending_verification';

ALTER TABLE users
    ADD CONSTRAINT chk_users_role
    CHECK (role IN ('admin', 'manager', 'member'));

DROP INDEX IF EXISTS idx_user_status_master_active_sort;
DROP TABLE IF EXISTS user_status_master;

DROP INDEX IF EXISTS idx_user_role_master_active_sort;
DROP TABLE IF EXISTS user_role_master;

-- Keep backup table for audit/manual rollback.

-- +goose StatementEnd
