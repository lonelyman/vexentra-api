package pgproject

import (
	"time"

	"vexentra-api/internal/modules/project"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectRoleMasterModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Code        string    `gorm:"column:code;not null"`
	NameTH      string    `gorm:"column:name_th;not null"`
	NameEN      string    `gorm:"column:name_en;not null"`
	Description *string   `gorm:"column:description"`
	SortOrder   int       `gorm:"column:sort_order;not null;default:0"`
	IsActive    bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt
}

func (projectRoleMasterModel) TableName() string { return "project_role_master" }

func (m *projectRoleMasterModel) ToEntity() *project.ProjectRole {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		t := m.DeletedAt.Time
		deletedAt = &t
	}
	return &project.ProjectRole{
		ID:          m.ID,
		Code:        m.Code,
		NameTH:      m.NameTH,
		NameEN:      m.NameEN,
		Description: m.Description,
		SortOrder:   m.SortOrder,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   deletedAt,
	}
}

type projectMemberRoleAssignmentModel struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProjectMemberID  uuid.UUID `gorm:"type:uuid;column:project_member_id;not null;index"`
	RoleID           uuid.UUID `gorm:"type:uuid;column:role_id;not null;index"`
	IsPrimary        bool      `gorm:"column:is_primary;not null;default:false"`
	AssignedByUserID uuid.UUID `gorm:"type:uuid;column:assigned_by_user_id;not null"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
	DeletedAt        gorm.DeletedAt
}

func (projectMemberRoleAssignmentModel) TableName() string { return "project_member_role_assignments" }
