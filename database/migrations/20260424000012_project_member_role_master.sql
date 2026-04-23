-- +goose Up
-- +goose StatementBegin

-- Master role definitions for per-project responsibilities.
CREATE TABLE project_role_master (
    id            UUID PRIMARY KEY,
    code          TEXT NOT NULL,
    name_th       TEXT NOT NULL,
    name_en       TEXT NOT NULL,
    description   TEXT,
    sort_order    INT NOT NULL DEFAULT 0,
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ,

    CONSTRAINT chk_project_role_master_code_format
        CHECK (code = lower(code) AND code ~ '^[a-z0-9_]+$')
);

CREATE UNIQUE INDEX idx_project_role_master_code_active
    ON project_role_master (code)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_project_role_master_active_sort
    ON project_role_master (is_active, sort_order)
    WHERE deleted_at IS NULL;

-- Role assignments on each project membership.
CREATE TABLE project_member_role_assignments (
    id                 UUID PRIMARY KEY,
    project_member_id  UUID NOT NULL REFERENCES project_members(id) ON DELETE CASCADE,
    role_id            UUID NOT NULL REFERENCES project_role_master(id) ON DELETE RESTRICT,
    is_primary         BOOLEAN NOT NULL DEFAULT false,
    assigned_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at         TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_project_member_roles_unique_active
    ON project_member_role_assignments (project_member_id, role_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_project_member_roles_primary_active
    ON project_member_role_assignments (project_member_id)
    WHERE deleted_at IS NULL AND is_primary = true;

CREATE INDEX idx_project_member_roles_member
    ON project_member_role_assignments (project_member_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_project_member_roles_role
    ON project_member_role_assignments (role_id)
    WHERE deleted_at IS NULL;

INSERT INTO project_role_master (id, code, name_th, name_en, description, sort_order, is_active) VALUES
    (gen_random_uuid(), 'project_manager', 'ผู้จัดการโครงการ', 'Project Manager', 'รับผิดชอบภาพรวมแผนงานและการส่งมอบ', 10, true),
    (gen_random_uuid(), 'solution_architect', 'สถาปนิกโซลูชัน', 'Solution Architect', 'ออกแบบสถาปัตยกรรมและทางเลือกเชิงเทคนิค', 20, true),
    (gen_random_uuid(), 'tech_lead', 'หัวหน้าทีมเทคนิค', 'Tech Lead', 'กำกับคุณภาพโค้ดและทิศทางการพัฒนา', 30, true),
    (gen_random_uuid(), 'backend_developer', 'นักพัฒนา Backend', 'Backend Developer', 'พัฒนา API และระบบหลังบ้าน', 40, true),
    (gen_random_uuid(), 'frontend_developer', 'นักพัฒนา Frontend', 'Frontend Developer', 'พัฒนาหน้าจอและประสบการณ์ผู้ใช้', 50, true),
    (gen_random_uuid(), 'qa_engineer', 'วิศวกรทดสอบ', 'QA Engineer', 'วางแผนและดำเนินการทดสอบคุณภาพระบบ', 60, true),
    (gen_random_uuid(), 'devops_engineer', 'วิศวกร DevOps', 'DevOps Engineer', 'ดูแล CI/CD โครงสร้างพื้นฐานและการ deploy', 70, true),
    (gen_random_uuid(), 'business_analyst', 'นักวิเคราะห์ธุรกิจ', 'Business Analyst', 'เก็บ requirement และแปลงเป็นงานเชิงระบบ', 80, true);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS project_member_role_assignments;
DROP TABLE IF EXISTS project_role_master;

-- +goose StatementEnd
