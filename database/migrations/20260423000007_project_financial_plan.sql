-- +goose Up
-- +goose StatementBegin

CREATE TABLE project_financial_plans (
    project_id             UUID PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
    contract_amount        NUMERIC(15,2) NOT NULL DEFAULT 0,
    retention_amount       NUMERIC(15,2) NOT NULL DEFAULT 0,
    planned_delivery_date  DATE,
    payment_note           TEXT,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_project_financial_contract_nonneg CHECK (contract_amount >= 0),
    CONSTRAINT chk_project_financial_retention_nonneg CHECK (retention_amount >= 0),
    CONSTRAINT chk_project_financial_retention_lte_contract CHECK (retention_amount <= contract_amount)
);

CREATE TABLE project_payment_installments (
    id                      UUID PRIMARY KEY,
    project_id              UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sort_order              INT NOT NULL DEFAULT 1,
    title                   TEXT NOT NULL,
    amount                  NUMERIC(15,2) NOT NULL,
    planned_delivery_date   DATE,
    planned_receive_date    DATE,
    note                    TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_project_installments_sort_order CHECK (sort_order > 0),
    CONSTRAINT chk_project_installments_amount_positive CHECK (amount > 0)
);

CREATE INDEX idx_project_installments_project_order
    ON project_payment_installments (project_id, sort_order ASC, created_at ASC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS project_payment_installments;
DROP TABLE IF EXISTS project_financial_plans;

-- +goose StatementEnd
