-- +goose Up
-- +goose StatementBegin
UPDATE project_statuses
SET label_th = 'ประมูล/เสนอราคา',
    updated_at = now()
WHERE status = 'bidding';

UPDATE project_statuses
SET label_th = 'ดำเนินการ',
    updated_at = now()
WHERE status = 'active';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE project_statuses
SET label_th = 'กำลังประมูล/เสนอราคา',
    updated_at = now()
WHERE status = 'bidding';

UPDATE project_statuses
SET label_th = 'กำลังดำเนินการ',
    updated_at = now()
WHERE status = 'active';
-- +goose StatementEnd