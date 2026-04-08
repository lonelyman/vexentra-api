// vexentra-api/pkg/custom_errors/errors.go
package custom_errors

import "net/http"

// AppError คือโครงสร้าง Error หลักของระบบ Vexentra
type AppError struct {
	HTTPStatus int         `json:"-"`       // ไว้ให้ Mapper ในชั้น Transport เลือกใช้ Status Code
	Code       string      `json:"code"`    // รหัส Error เช่น "ORDER_NOT_FOUND"
	Message    string      `json:"message"` // ข้อความสำหรับ User
	Details    interface{} `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

// New เป็น Constructor หลักสำหรับสร้าง AppError
func New(status int, code, message string, details ...interface{}) *AppError {
	var d interface{}
	if len(details) > 0 {
		d = details[0]
	}
	return &AppError{
		HTTPStatus: status,
		Code:       code,
		Message:    message,
		Details:    d,
	}
}

// ───────────────────────────────────────────────────────────────────────
//  Standard Error Codes (ยกมาจากของเก่านายท่าน)
// ───────────────────────────────────────────────────────────────────────

const (
	// Auth & Permission
	ErrUnauthorized = "UNAUTHORIZED"
	ErrForbidden    = "FORBIDDEN"

	// Validation
	ErrValidation    = "VALIDATION_ERROR"
	ErrInvalidFormat = "INVALID_FORMAT"
	ErrMissingField  = "MISSING_FIELD"

	// Resource
	ErrNotFound      = "NOT_FOUND"
	ErrAlreadyExists = "ALREADY_EXISTS"

	// System
	ErrInternal = "INTERNAL_SERVER_ERROR"
)

// ───────────────────────────────────────────────────────────────────────
//  Helper Functions (สำหรับ Service Layer เรียกใช้ได้สะดวก)
// ───────────────────────────────────────────────────────────────────────

func NewNotFoundError(code, message string) *AppError {
	return New(http.StatusNotFound, code, message)
}

func NewBadRequestError(code, message string, details ...interface{}) *AppError {
	return New(http.StatusBadRequest, code, message, details...)
}

func NewInternalError(message string) *AppError {
	return New(http.StatusInternalServerError, ErrInternal, message)
}
