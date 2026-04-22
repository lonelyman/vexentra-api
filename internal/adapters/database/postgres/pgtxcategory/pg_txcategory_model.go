package pgtxcategory

import (
	"time"

	"vexentra-api/internal/modules/txcategory"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type txCategoryModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Code      string    `gorm:"not null"`
	Name      string    `gorm:"not null"`
	Type      string    `gorm:"type:transaction_type;not null"`
	IconKey   *string   `gorm:"column:icon_key"`
	IsSystem  bool      `gorm:"column:is_system;not null;default:false"`
	IsActive  bool      `gorm:"column:is_active;not null;default:true"`
	SortOrder int       `gorm:"column:sort_order;not null;default:0"`

	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (txCategoryModel) TableName() string { return "transaction_categories" }

func (m *txCategoryModel) ToEntity() *txcategory.TransactionCategory {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		t := m.DeletedAt.Time
		deletedAt = &t
	}
	return &txcategory.TransactionCategory{
		ID:        m.ID,
		Code:      m.Code,
		Name:      m.Name,
		Type:      txcategory.TransactionType(m.Type),
		IconKey:   m.IconKey,
		IsSystem:  m.IsSystem,
		IsActive:  m.IsActive,
		SortOrder: m.SortOrder,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

func fromCategory(c *txcategory.TransactionCategory) *txCategoryModel {
	return &txCategoryModel{
		ID:        c.ID,
		Code:      c.Code,
		Name:      c.Name,
		Type:      string(c.Type),
		IconKey:   c.IconKey,
		IsSystem:  c.IsSystem,
		IsActive:  c.IsActive,
		SortOrder: c.SortOrder,
	}
}
