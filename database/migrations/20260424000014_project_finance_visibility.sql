-- +goose Up
-- +goose StatementBegin

ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS contract_finance_visibility TEXT NOT NULL DEFAULT 'all_members',
    ADD COLUMN IF NOT EXISTS expense_finance_visibility  TEXT NOT NULL DEFAULT 'all_members';

ALTER TABLE projects
    DROP CONSTRAINT IF EXISTS chk_projects_contract_finance_visibility,
    DROP CONSTRAINT IF EXISTS chk_projects_expense_finance_visibility;

ALTER TABLE projects
    ADD CONSTRAINT chk_projects_contract_finance_visibility
        CHECK (contract_finance_visibility IN ('all_members', 'lead_coordinator_staff', 'staff_only')),
    ADD CONSTRAINT chk_projects_expense_finance_visibility
        CHECK (expense_finance_visibility IN ('all_members', 'lead_coordinator_staff', 'staff_only'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE projects
    DROP CONSTRAINT IF EXISTS chk_projects_contract_finance_visibility,
    DROP CONSTRAINT IF EXISTS chk_projects_expense_finance_visibility;

ALTER TABLE projects
    DROP COLUMN IF EXISTS contract_finance_visibility,
    DROP COLUMN IF EXISTS expense_finance_visibility;

-- +goose StatementEnd
