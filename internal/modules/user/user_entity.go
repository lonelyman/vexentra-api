package user

import (
	"time"

	"github.com/google/uuid"
)

// สถานะบัญชีผู้ใช้
const (
	UserStatusActive              = "active"
	UserStatusBanned              = "banned"
	UserStatusPendingVerification = "pending_verification"
)

// Role ของผู้ใช้ในระบบ
const (
	UserRoleUser  = "user"
	UserRoleAdmin = "admin"
)

// ประเภทผู้ให้บริการยืนยันตัวตน
const (
	AuthProviderLocal  = "local"
	AuthProviderGoogle = "google"
	AuthProviderGitHub = "github"
)

// User คือ domain entity หลัก ไม่มี JSON tag —
// การ serialize ทำที่ transport layer (UserResponse) เท่านั้น
type User struct {
	ID          uuid.UUID
	PersonID    uuid.UUID // FK → persons.id (linked identity record)
	Username    string
	Email       string
	Role        string     // user | admin
	Status      string     // active | banned | pending_verification
	LastLoginAt *time.Time // nil = ยังไม่เคย login
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time // nil = active; non-nil = soft deleted

	// Email Verification
	IsEmailVerified                 bool
	EmailVerificationToken          *string
	EmailVerificationTokenExpiresAt *time.Time

	// Password Reset
	PasswordResetToken          *string
	PasswordResetTokenExpiresAt *time.Time

	// Association — โหลดเมื่อต้องการเท่านั้น
	Auths []*UserAuth
}

// UserAuth เก็บข้อมูลการยืนยันตัวตนแยกต่างหาก
// รองรับหลาย provider ต่อ 1 user (local, google, github ฯลฯ)
type UserAuth struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Provider     string  // "local" | "google" | "github"
	ProviderID   string  // OAuth: ID จาก provider; local: ใช้ email
	Secret       string  // hashed password (local only); ว่างสำหรับ OAuth
	RefreshToken *string // refresh token ล่าสุดที่ออกให้ user นี้
}
