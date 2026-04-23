-- +goose Up
-- +goose StatementBegin

INSERT INTO project_role_master (id, code, name_th, name_en, description, sort_order, is_active)
SELECT
    gen_random_uuid(),
    'coordinator',
    'ผู้ประสานงานโครงการ',
    'Project Coordinator',
    'ประสานงานและดูแลการอัปเดตข้อมูลโครงการ',
    15,
    true
WHERE NOT EXISTS (
    SELECT 1
    FROM project_role_master
    WHERE code = 'coordinator' AND deleted_at IS NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

UPDATE project_role_master
SET deleted_at = now(), is_active = false, updated_at = now()
WHERE code = 'coordinator' AND deleted_at IS NULL;

-- +goose StatementEnd
