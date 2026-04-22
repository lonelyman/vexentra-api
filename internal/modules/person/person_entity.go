package person

import (
	"time"

	"github.com/google/uuid"
)

// Person คือ identity หลักของระบบ — แยกออกจาก User (auth account)
// คนที่ไม่ได้สมัครสมาชิกก็มี Person record ได้ (linked_user_id = nil)
// เมื่อสมัครและ claim แล้ว linked_user_id จะถูก set
type Person struct {
	ID                   uuid.UUID
	Name                 string
	InviteEmail          *string    // nullable — อีเมลที่ admin ระบุตอนสร้าง; ใช้ suggest linking ตอน Register
	InviteToken          *string    // one-time token สำหรับ invite link (auto-link เมื่อใช้)
	InviteTokenExpiresAt *time.Time // nil = ไม่มี token
	LinkedUserID         *uuid.UUID // nil = ยังไม่มี account; non-nil = เชื่อมกับ users.id
	CreatedByUserID      *uuid.UUID // user ที่สร้าง person record นี้ (nil = self-registered)
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            *time.Time
}
