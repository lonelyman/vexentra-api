-- +goose Up
-- Task management: per-project task list with status/priority/assignee

CREATE TABLE tasks (
    id                  UUID         NOT NULL PRIMARY KEY,
    project_id          UUID         NOT NULL REFERENCES projects(id)  ON DELETE RESTRICT,
    title               TEXT         NOT NULL,
    description         TEXT,
    status              VARCHAR(20)  NOT NULL DEFAULT 'todo'
        CHECK (status IN ('todo', 'in_progress', 'done', 'cancelled')),
    priority            VARCHAR(10)  NOT NULL DEFAULT 'medium'
        CHECK (priority IN ('low', 'medium', 'high')),
    assigned_person_id  UUID         REFERENCES persons(id) ON DELETE SET NULL,
    due_date            DATE,
    created_by_user_id  UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

-- Fast lookup of all non-deleted tasks for a project
CREATE INDEX idx_tasks_project_id ON tasks (project_id) WHERE deleted_at IS NULL;

-- Filter by project + status (most common query pattern)
CREATE INDEX idx_tasks_project_status ON tasks (project_id, status) WHERE deleted_at IS NULL;

-- My tasks view (filter by assignee)
CREATE INDEX idx_tasks_assigned_person ON tasks (assigned_person_id) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS tasks;
