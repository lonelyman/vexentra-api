package pguser

import (
	"time"
	"vexentra-api/internal/modules/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────────────────────────────────
//  userModel — ตาราง users
// ─────────────────────────────────────────────────────────────────────────────

type userModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	PersonID    uuid.UUID  `gorm:"type:uuid;column:person_id;not null;uniqueIndex:idx_users_person,where:deleted_at IS NULL"`
	Username    string     `gorm:"size:50;not null"`
	Email       string     `gorm:"size:254;uniqueIndex:idx_users_email_active,where:deleted_at IS NULL;not null"`
	Role        string     `gorm:"size:20;column:role;not null;default:'user'"`
	Status      string     `gorm:"size:30;column:status;not null;default:'pending_verification'"`
	LastLoginAt *time.Time `gorm:"column:last_login_at"`

	// Email Verification
	IsEmailVerified                 bool       `gorm:"column:is_email_verified;not null;default:false"`
	EmailVerificationToken          *string    `gorm:"size:255;column:email_verification_token;uniqueIndex:idx_users_email_verification_token_active,where:deleted_at IS NULL"`
	EmailVerificationTokenExpiresAt *time.Time `gorm:"column:email_verification_token_expires_at"`

	// Password Reset
	PasswordResetToken          *string    `gorm:"size:255;column:password_reset_token;uniqueIndex:idx_users_password_reset_token_active,where:deleted_at IS NULL"`
	PasswordResetTokenExpiresAt *time.Time `gorm:"column:password_reset_token_expires_at"`
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
	DeletedAt                   gorm.DeletedAt `gorm:"index"` // soft delete

	// Association
	Auths []userAuthModel `gorm:"foreignKey:UserID"`
}

func (userModel) TableName() string { return "users" }

func (m *userModel) ToEntity() *user.User {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		t := m.DeletedAt.Time
		deletedAt = &t
	}
	return &user.User{
		ID:          m.ID,
		PersonID:    m.PersonID,
		Username:    m.Username,
		Email:       m.Email,
		Role:        m.Role,
		Status:      m.Status,
		LastLoginAt: m.LastLoginAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   deletedAt,

		IsEmailVerified:                 m.IsEmailVerified,
		EmailVerificationToken:          m.EmailVerificationToken,
		EmailVerificationTokenExpiresAt: m.EmailVerificationTokenExpiresAt,
		PasswordResetToken:              m.PasswordResetToken,
		PasswordResetTokenExpiresAt:     m.PasswordResetTokenExpiresAt,
	}
}

func fromUserEntity(u *user.User) *userModel {
	return &userModel{
		ID:       u.ID,
		PersonID: u.PersonID,
		Username: u.Username,
		Email:    u.Email,
		Role:     u.Role,
		Status:   u.Status,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
//  userAuthModel — ตาราง user_auths
// ─────────────────────────────────────────────────────────────────────────────

type userAuthModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index"`
	Provider     string    `gorm:"size:30;not null"`                              // local | google | github
	ProviderID   string    `gorm:"size:254;uniqueIndex:idx_provider_provider_id"` // email สำหรับ local; user_id จาก OAuth
	Secret       string    `gorm:"size:60;column:secret"`                         // bcrypt hash (60 ตัวเสมอ)
	RefreshToken *string   `gorm:"size:512;column:refresh_token"`                 // refresh token ล่าสุด
}

func (userAuthModel) TableName() string { return "user_auths" }

func (m *userAuthModel) ToEntity() *user.UserAuth {
	return &user.UserAuth{
		ID:           m.ID,
		UserID:       m.UserID,
		Provider:     m.Provider,
		ProviderID:   m.ProviderID,
		Secret:       m.Secret,
		RefreshToken: m.RefreshToken,
	}
}

func fromUserAuthEntity(a *user.UserAuth) *userAuthModel {
	return &userAuthModel{
		ID:           a.ID,
		UserID:       a.UserID,
		Provider:     a.Provider,
		ProviderID:   a.ProviderID,
		Secret:       a.Secret,
		RefreshToken: a.RefreshToken,
	}
}
