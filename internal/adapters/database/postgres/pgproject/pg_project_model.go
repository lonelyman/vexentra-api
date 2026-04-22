package pgproject

import (
	"time"

	"vexentra-api/internal/modules/project"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// projectModel mirrors the `projects` table (migration 20260422000003).
// GORM tags describe columns for query purposes only — DDL is goose-managed.
type projectModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProjectCode string    `gorm:"column:project_code;not null;uniqueIndex"`

	Name        string  `gorm:"not null"`
	Description *string `gorm:"column:description"`

	Status        string  `gorm:"type:project_status;not null;default:'draft'"`
	ClosureReason *string `gorm:"type:project_closure_reason;column:closure_reason"`

	ClientPersonID *uuid.UUID `gorm:"type:uuid;column:client_person_id"`
	ClientNameRaw  *string    `gorm:"column:client_name_raw"`
	ClientEmailRaw *string    `gorm:"column:client_email_raw"`

	ScheduledStartAt *time.Time `gorm:"column:scheduled_start_at"`
	DeadlineAt       *time.Time `gorm:"column:deadline_at"`
	ActivatedAt      *time.Time `gorm:"column:activated_at"`
	ClosedAt         *time.Time `gorm:"column:closed_at"`

	CreatedByUserID uuid.UUID      `gorm:"type:uuid;column:created_by_user_id;not null"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (projectModel) TableName() string { return "projects" }

func (m *projectModel) ToEntity() *project.Project {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		t := m.DeletedAt.Time
		deletedAt = &t
	}
	var closureReason *project.ProjectClosureReason
	if m.ClosureReason != nil {
		v := project.ProjectClosureReason(*m.ClosureReason)
		closureReason = &v
	}
	return &project.Project{
		ID:               m.ID,
		ProjectCode:      m.ProjectCode,
		Name:             m.Name,
		Description:      m.Description,
		Status:           project.ProjectStatus(m.Status),
		ClosureReason:    closureReason,
		ClientPersonID:   m.ClientPersonID,
		ClientNameRaw:    m.ClientNameRaw,
		ClientEmailRaw:   m.ClientEmailRaw,
		ScheduledStartAt: m.ScheduledStartAt,
		DeadlineAt:       m.DeadlineAt,
		ActivatedAt:      m.ActivatedAt,
		ClosedAt:         m.ClosedAt,
		CreatedByUserID:  m.CreatedByUserID,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
		DeletedAt:        deletedAt,
	}
}

func fromProject(p *project.Project) *projectModel {
	var closureReason *string
	if p.ClosureReason != nil {
		s := string(*p.ClosureReason)
		closureReason = &s
	}
	return &projectModel{
		ID:               p.ID,
		ProjectCode:      p.ProjectCode,
		Name:             p.Name,
		Description:      p.Description,
		Status:           string(p.Status),
		ClosureReason:    closureReason,
		ClientPersonID:   p.ClientPersonID,
		ClientNameRaw:    p.ClientNameRaw,
		ClientEmailRaw:   p.ClientEmailRaw,
		ScheduledStartAt: p.ScheduledStartAt,
		DeadlineAt:       p.DeadlineAt,
		ActivatedAt:      p.ActivatedAt,
		ClosedAt:         p.ClosedAt,
		CreatedByUserID:  p.CreatedByUserID,
	}
}
