-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS project_statuses (
    status          project_status PRIMARY KEY,
    label_th        TEXT NOT NULL,
    phase           TEXT NOT NULL CHECK (phase IN ('backlog', 'pre_sales', 'delivery', 'terminal')),
    sort_order      SMALLINT NOT NULL,
    is_terminal     BOOLEAN NOT NULL DEFAULT FALSE,
    requires_client BOOLEAN NOT NULL DEFAULT FALSE,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_project_statuses_active_sort
ON project_statuses (is_active, sort_order);

INSERT INTO project_statuses (status, label_th, phase, sort_order, is_terminal, requires_client, is_active)
VALUES
    ('draft',   'แบบร่าง',             'backlog',   10, FALSE, FALSE, TRUE),
    ('planned', 'วางแผน',               'pre_sales', 20, FALSE, FALSE, TRUE),
    ('bidding', 'กำลังประมูล/เสนอราคา',   'pre_sales', 30, FALSE, FALSE, TRUE),
    ('active',  'กำลังดำเนินการ',        'delivery',  40, FALSE, TRUE,  TRUE),
    ('on_hold', 'ระงับชั่วคราว',          'delivery',  50, FALSE, TRUE,  TRUE),
    ('closed',  'ปิดโครงการ',            'terminal',  60, TRUE,  TRUE,  TRUE)
ON CONFLICT (status) DO UPDATE
SET
    label_th = EXCLUDED.label_th,
    phase = EXCLUDED.phase,
    sort_order = EXCLUDED.sort_order,
    is_terminal = EXCLUDED.is_terminal,
    requires_client = EXCLUDED.requires_client,
    is_active = EXCLUDED.is_active,
    updated_at = now();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS project_statuses;
-- +goose StatementEnd
