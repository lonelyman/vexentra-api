package pgproject

import (
	"time"

	"vexentra-api/internal/modules/project"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// projectTransactionModel mirrors the `project_transactions` table.
// Amount maps to NUMERIC(15,2) via shopspring/decimal's Scanner/Valuer.
type projectTransactionModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProjectID  uuid.UUID `gorm:"type:uuid;column:project_id;not null;index"`
	CategoryID uuid.UUID `gorm:"type:uuid;column:category_id;not null;index"`

	Amount       decimal.Decimal `gorm:"type:numeric(15,2);not null"`
	CurrencyCode string          `gorm:"type:char(3);column:currency_code;not null;default:'THB'"`
	Note         *string         `gorm:"column:note"`
	OccurredAt   time.Time       `gorm:"column:occurred_at;not null"`

	CreatedByUserID uuid.UUID      `gorm:"type:uuid;column:created_by_user_id;not null"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (projectTransactionModel) TableName() string { return "project_transactions" }

func (m *projectTransactionModel) ToEntity() *project.ProjectTransaction {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		t := m.DeletedAt.Time
		deletedAt = &t
	}
	return &project.ProjectTransaction{
		ID:              m.ID,
		ProjectID:       m.ProjectID,
		CategoryID:      m.CategoryID,
		Amount:          m.Amount,
		CurrencyCode:    m.CurrencyCode,
		Note:            m.Note,
		OccurredAt:      m.OccurredAt,
		CreatedByUserID: m.CreatedByUserID,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
		DeletedAt:       deletedAt,
	}
}

func fromTransaction(t *project.ProjectTransaction) *projectTransactionModel {
	return &projectTransactionModel{
		ID:              t.ID,
		ProjectID:       t.ProjectID,
		CategoryID:      t.CategoryID,
		Amount:          t.Amount,
		CurrencyCode:    t.CurrencyCode,
		Note:            t.Note,
		OccurredAt:      t.OccurredAt,
		CreatedByUserID: t.CreatedByUserID,
	}
}
