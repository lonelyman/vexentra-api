-- +goose Up
-- +goose StatementBegin

ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS project_kind TEXT NOT NULL DEFAULT 'client_delivery';

ALTER TABLE projects
    DROP CONSTRAINT IF EXISTS chk_projects_project_kind;

ALTER TABLE projects
    ADD CONSTRAINT chk_projects_project_kind
        CHECK (project_kind IN ('client_delivery', 'internal_continuous'));

ALTER TABLE projects
    DROP CONSTRAINT IF EXISTS chk_projects_client_required_when_active;

ALTER TABLE projects
    ADD CONSTRAINT chk_projects_client_required_when_active
        CHECK (
            project_kind = 'internal_continuous'
            OR status IN ('draft', 'planned', 'bidding')
            OR client_person_id IS NOT NULL
        );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE projects
    DROP CONSTRAINT IF EXISTS chk_projects_client_required_when_active;

ALTER TABLE projects
    ADD CONSTRAINT chk_projects_client_required_when_active
        CHECK (
            status IN ('draft', 'planned', 'bidding')
            OR client_person_id IS NOT NULL
        );

ALTER TABLE projects
    DROP CONSTRAINT IF EXISTS chk_projects_project_kind;

ALTER TABLE projects
    DROP COLUMN IF EXISTS project_kind;

-- +goose StatementEnd
