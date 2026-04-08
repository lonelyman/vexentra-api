// vexentra-api/internal/transport/http/presenter/presenter.go
package presenter

import (
	"math"
	"vexentra-api/pkg/custom_errors"

	"github.com/gofiber/fiber/v3"
)

// ───────────────────────────────────────────────────────────────────────
//  Models & Structures
// ───────────────────────────────────────────────────────────────────────

// Pagination เก็บข้อมูลสำหรับการแบ่งหน้าแบบ Offset-based
type Pagination struct {
	TotalRecords *int    `json:"total_records,omitempty"`
	Limit        *int    `json:"limit,omitempty"`
	Offset       *int    `json:"offset,omitempty"`
	TotalPages   *int    `json:"total_pages,omitempty"`
	CurrentPage  *int    `json:"current_page,omitempty"`
	NextCursor   *string `json:"next_cursor,omitempty"` // รองรับเผื่ออนาคตใช้ Hybrid
	HasMore      *bool   `json:"has_more,omitempty"`
}

// SuccessResponse คือ Envelope หลักสำหรับ Response ที่สำเร็จ
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// ListPayload คือโครงสร้างภายใน data สำหรับข้อมูลที่เป็นรายการ (List/Collection)
type ListPayload struct {
	Items      interface{} `json:"items"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// ErrorResponse คือ Envelope หลักสำหรับ Response ที่เกิดข้อผิดพลาด
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail รายละเอียดของ Error ที่จะส่งกลับไป
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ───────────────────────────────────────────────────────────────────────
//  Render Functions (The Strategic Presenters)
// ───────────────────────────────────────────────────────────────────────

// RenderList สำหรับส่งข้อมูลชุด (Collection) พร้อม Pagination ภายใต้ data envelope
// ใช้ Status 200 OK เป็นค่าเริ่มต้น
func RenderList(c fiber.Ctx, items interface{}, pagination ...*Pagination) error {
	var pg *Pagination
	if len(pagination) > 0 {
		pg = pagination[0]
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Data: ListPayload{
			Items:      items,
			Pagination: pg,
		},
	})
}

// RenderItem สำหรับส่งข้อมูลชิ้นเดียว (Object/Item) ภายใต้ data envelope
// สามารถระบุ statusCode เพิ่มเติมได้ (เช่น fiber.StatusCreated สำหรับการ POST สำเร็จ)
func RenderItem(c fiber.Ctx, data interface{}, statusCode ...int) error {
	status := fiber.StatusOK
	if len(statusCode) > 0 {
		status = statusCode[0]
	}

	return c.Status(status).JSON(SuccessResponse{
		Data: data,
	})
}

// RenderError จัดการส่ง Error Response ตามมาตรฐานของ AppError
func RenderError(c fiber.Ctx, err error) error {
	if appErr, ok := err.(*custom_errors.AppError); ok {
		return c.Status(appErr.HTTPStatus).JSON(ErrorResponse{
			Error: ErrorDetail{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			},
		})
	}

	// กรณี Error อื่นๆ ที่ไม่ได้นิยามไว้ (Internal Server Error)
	return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		Error: ErrorDetail{
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "An unexpected error occurred",
		},
	})
}

// ───────────────────────────────────────────────────────────────────────
//  Helpers
// ───────────────────────────────────────────────────────────────────────

// NewOffsetPagination คำนวณ Metadata สำหรับการทำ Pagination แบบ Page/Limit
func NewOffsetPagination(totalRecords, limit, offset int) *Pagination {
	if limit <= 0 {
		limit = 10
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(limit)))
	currentPage := int(math.Floor(float64(offset)/float64(limit))) + 1
	hasMore := offset+limit < totalRecords

	return &Pagination{
		TotalRecords: &totalRecords,
		Limit:        &limit,
		Offset:       &offset,
		TotalPages:   &totalPages,
		CurrentPage:  &currentPage,
		HasMore:      &hasMore,
	}
}
