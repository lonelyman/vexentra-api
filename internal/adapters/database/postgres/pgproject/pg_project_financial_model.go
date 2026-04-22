package pgproject

import (
	"time"

	"vexentra-api/internal/modules/project"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type projectFinancialPlanModel struct {
	ProjectID           uuid.UUID       `gorm:"type:uuid;primaryKey;column:project_id"`
	ContractAmount      decimal.Decimal `gorm:"type:numeric(15,2);column:contract_amount"`
	RetentionAmount     decimal.Decimal `gorm:"type:numeric(15,2);column:retention_amount"`
	PlannedDeliveryDate *time.Time      `gorm:"column:planned_delivery_date"`
	PaymentNote         *string         `gorm:"column:payment_note"`
	CreatedAt           time.Time       `gorm:"column:created_at"`
	UpdatedAt           time.Time       `gorm:"column:updated_at"`
}

func (projectFinancialPlanModel) TableName() string { return "project_financial_plans" }

type projectPaymentInstallmentModel struct {
	ID                  uuid.UUID       `gorm:"type:uuid;primaryKey"`
	ProjectID           uuid.UUID       `gorm:"type:uuid;column:project_id"`
	SortOrder           int             `gorm:"column:sort_order"`
	Title               string          `gorm:"column:title"`
	Amount              decimal.Decimal `gorm:"type:numeric(15,2);column:amount"`
	PlannedDeliveryDate *time.Time      `gorm:"column:planned_delivery_date"`
	PlannedReceiveDate  *time.Time      `gorm:"column:planned_receive_date"`
	Note                *string         `gorm:"column:note"`
	CreatedAt           time.Time       `gorm:"column:created_at"`
	UpdatedAt           time.Time       `gorm:"column:updated_at"`
}

func (projectPaymentInstallmentModel) TableName() string { return "project_payment_installments" }

func (m *projectPaymentInstallmentModel) ToEntity() *project.ProjectPaymentInstallment {
	return &project.ProjectPaymentInstallment{
		ID:                  m.ID,
		ProjectID:           m.ProjectID,
		SortOrder:           m.SortOrder,
		Title:               m.Title,
		Amount:              m.Amount,
		PlannedDeliveryDate: m.PlannedDeliveryDate,
		PlannedReceiveDate:  m.PlannedReceiveDate,
		Note:                m.Note,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}
