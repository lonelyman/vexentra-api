package pgperson

import (
	"time"
	"vexentra-api/internal/modules/person"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// personModel — ตาราง persons
// เป็น identity หลักของระบบ แยกจาก users (auth account)
type personModel struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Name                 string     `gorm:"size:200;not null"`
	InviteEmail          *string    `gorm:"column:invite_email;size:254;uniqueIndex:idx_persons_invite_email_active,where:deleted_at IS NULL"`
	InviteToken          *string    `gorm:"column:invite_token;size:128;uniqueIndex:idx_persons_invite_token,where:deleted_at IS NULL"`
	InviteTokenExpiresAt *time.Time `gorm:"column:invite_token_expires_at"`
	LinkedUserID         *uuid.UUID `gorm:"type:uuid;column:linked_user_id;uniqueIndex:idx_persons_linked_user,where:deleted_at IS NULL"`
	CreatedByUserID      uuid.UUID  `gorm:"type:uuid;column:created_by_user_id;not null;index"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            gorm.DeletedAt `gorm:"index"`
}

func (personModel) TableName() string { return "persons" }

func (m *personModel) ToEntity() *person.Person {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		t := m.DeletedAt.Time
		deletedAt = &t
	}
	return &person.Person{
		ID:                   m.ID,
		Name:                 m.Name,
		InviteEmail:          m.InviteEmail,
		InviteToken:          m.InviteToken,
		InviteTokenExpiresAt: m.InviteTokenExpiresAt,
		LinkedUserID:         m.LinkedUserID,
		CreatedByUserID:      m.CreatedByUserID,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
		DeletedAt:            deletedAt,
	}
}

func fromEntity(p *person.Person) *personModel {
	return &personModel{
		ID:                   p.ID,
		Name:                 p.Name,
		InviteEmail:          p.InviteEmail,
		InviteToken:          p.InviteToken,
		InviteTokenExpiresAt: p.InviteTokenExpiresAt,
		LinkedUserID:         p.LinkedUserID,
		CreatedByUserID:      p.CreatedByUserID,
	}
}
