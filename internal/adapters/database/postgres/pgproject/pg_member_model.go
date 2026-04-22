package pgproject

import (
	"time"

	"vexentra-api/internal/modules/project"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// projectMemberModel mirrors the `project_members` table (migration 20260422000003).
// Structure is flat + is_lead — no per-member role column. Joined date is CreatedAt,
// left date is DeletedAt (soft delete).
type projectMemberModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProjectID uuid.UUID `gorm:"type:uuid;column:project_id;not null;index"`
	PersonID  uuid.UUID `gorm:"type:uuid;column:person_id;not null;index"`

	IsLead bool `gorm:"column:is_lead;not null;default:false"`

	AddedByUserID uuid.UUID      `gorm:"type:uuid;column:added_by_user_id;not null"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (projectMemberModel) TableName() string { return "project_members" }

func (m *projectMemberModel) ToEntity() *project.ProjectMember {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		t := m.DeletedAt.Time
		deletedAt = &t
	}
	return &project.ProjectMember{
		ID:            m.ID,
		ProjectID:     m.ProjectID,
		PersonID:      m.PersonID,
		IsLead:        m.IsLead,
		AddedByUserID: m.AddedByUserID,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
		DeletedAt:     deletedAt,
	}
}

func fromMember(m *project.ProjectMember) *projectMemberModel {
	return &projectMemberModel{
		ID:            m.ID,
		ProjectID:     m.ProjectID,
		PersonID:      m.PersonID,
		IsLead:        m.IsLead,
		AddedByUserID: m.AddedByUserID,
	}
}
