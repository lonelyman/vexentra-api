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

// Role ของผู้ใช้ในระบบ — ตรงกับ CHECK constraint ใน migration 20260422000003
// (legacy value "user" ถูก migrate เป็น "member" แล้ว)
const (
	UserRoleMember  = "member"
	UserRoleManager = "manager"
	UserRoleAdmin   = "admin"
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
	Role        string     // member | manager | admin
	Status      string     // active | banned | pending_verification
	LastLoginAt *time.Time // nil = ยังไม่เคย login
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time // nil = active; non-nil = soft deleted
	// Password policy flags
	ForcePasswordChange bool
	PasswordChangedAt   *time.Time

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

// Caller identifies the acting user for permission-sensitive service calls.
// Handlers build it from JWT claims and pass it explicitly — services never
// read identity from ctx or globals.
type Caller struct {
	UserID   uuid.UUID
	PersonID uuid.UUID
	Role     string // UserRoleMember | UserRoleManager | UserRoleAdmin
}

func (c Caller) IsAdmin() bool   { return c.Role == UserRoleAdmin }
func (c Caller) IsManager() bool { return c.Role == UserRoleManager }

// IsStaff returns true for roles that bypass ownership/membership checks
// (admin + manager). Members are not staff.
func (c Caller) IsStaff() bool { return c.IsAdmin() || c.IsManager() }

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

// UserRoleMaster is a master-data row for account roles.
type UserRoleMaster struct {
	Code      string
	LabelTH   string
	LabelEN   string
	SortOrder int
	IsActive  bool
}

// UserStatusMaster is a master-data row for account statuses.
type UserStatusMaster struct {
	Code      string
	LabelTH   string
	LabelEN   string
	SortOrder int
	IsActive  bool
}
