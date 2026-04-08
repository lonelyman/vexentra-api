package user

import "time"

// User คือ Pure Domain Entity สำหรับใช้งานใน Business Logic
type User struct {
	ID          uint      `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Password    string    `json:"-"` // <--- เพิ่มตรงนี้เพื่อความปลอดภัย (Security Gap Fix)
	DisplayName string    `json:"display_name"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
