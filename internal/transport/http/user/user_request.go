package userhdl

// RegisterRequest ใช้รับข้อมูลจาก Frontend พร้อม Tag สำหรับการตรวจสอบ
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=4,max=20"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}
