-- +goose Up
-- +goose StatementBegin

-- Add system expense category: penalty fee (ค่าปรับ)
-- Safe for repeated deploys: update existing active row if present, else insert.
UPDATE transaction_categories
SET
    name = 'ค่าปรับ',
    type = 'expense',
    icon_key = 'alert-triangle',
    is_system = true,
    is_active = true,
    sort_order = 80,
    updated_at = now(),
    deleted_at = NULL
WHERE code = 'penalty' AND deleted_at IS NULL;

INSERT INTO transaction_categories (id, code, name, type, icon_key, is_system, is_active, sort_order)
SELECT
    gen_random_uuid(),
    'penalty',
    'ค่าปรับ',
    'expense',
    'alert-triangle',
    true,
    true,
    80
WHERE NOT EXISTS (
    SELECT 1
    FROM transaction_categories
    WHERE code = 'penalty' AND deleted_at IS NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM transaction_categories
WHERE code = 'penalty' AND is_system = true;

-- +goose StatementEnd
