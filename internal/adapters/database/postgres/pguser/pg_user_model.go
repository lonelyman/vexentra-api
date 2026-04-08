package pguser

import (
	"time"
	"vexentra-api/internal/modules/user"
)

// userModel คือ Schema ที่ตรงกับ Table ใน Database
type userModel struct {
	ID           uint   `gorm:"primaryKey;autoIncrement"`
	Username     string `gorm:"uniqueIndex;not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"column:password_hash;not null"`
	DisplayName  string `gorm:"column:display_name"`
	IsActive     bool   `gorm:"column:is_active;default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (userModel) TableName() string { return "users" }

// Mapper: แปลง DB Model -> Domain Entity
func (m *userModel) ToEntity() *user.User {
	return &user.User{
		ID:          m.ID,
		Username:    m.Username,
		Email:       m.Email,
		Password:    m.PasswordHash,
		DisplayName: m.DisplayName,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// Mapper: แปลง Domain Entity -> DB Model (เพื่อเซฟลง DB)
func fromUserEntity(u *user.User) *userModel {
	return &userModel{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.Password,
		DisplayName:  u.DisplayName,
		IsActive:     u.IsActive,
	}
}
