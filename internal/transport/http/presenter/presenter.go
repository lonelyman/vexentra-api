// vexentra-api/internal/transport/http/presenter/presenter.go
package presenter

import (
	"math"
	"strconv"
	"vexentra-api/pkg/custom_errors"

	"github.com/gofiber/fiber/v3"
)

const (
	defaultLimit = 10
	maxLimit     = 100
)

// ─────────────────────────────────────────────────────────────────────
//  Query Parsers — call these in handlers to get validated params
// ─────────────────────────────────────────────────────────────────────

// OffsetQuery holds parsed page/limit query parameters.
type OffsetQuery struct {
	Page   int // 1-based
	Limit  int
	Offset int // computed: (Page-1) * Limit
}

// ParseOffsetQuery parses ?page=&limit= from the request.
// Defaults: page=1, limit=10. Max limit is capped at 100.
func ParseOffsetQuery(c fiber.Ctx) OffsetQuery {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.Query("limit", strconv.Itoa(defaultLimit)))
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	return OffsetQuery{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

// CursorQuery holds parsed cursor/limit query parameters.
type CursorQuery struct {
	Cursor string // empty = first page
	Limit  int
}

// ParseCursorQuery parses ?cursor=&limit= from the request.
// Defaults: cursor="" (first page), limit=10. Max limit is capped at 100.
func ParseCursorQuery(c fiber.Ctx) CursorQuery {
	limit, _ := strconv.Atoi(c.Query("limit", strconv.Itoa(defaultLimit)))
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	return CursorQuery{
		Cursor: c.Query("cursor", ""),
		Limit:  limit,
	}
}

// ─────────────────────────────────────────────────────────────────────
//  Pagination Metadata
// ─────────────────────────────────────────────────────────────────────

// OffsetPagination is the metadata envelope for page/limit-based lists.
// Use with RenderList when the client needs jump-to-page navigation.
type OffsetPagination struct {
	TotalRecords int  `json:"total_records"`
	TotalPages   int  `json:"total_pages"`
	CurrentPage  int  `json:"current_page"`
	PerPage      int  `json:"per_page"`
	HasMore      bool `json:"has_more"`
}

// NewOffsetPagination computes offset pagination metadata from raw values.
// Pass OffsetQuery.Offset and OffsetQuery.Limit to keep handlers concise.
func NewOffsetPagination(totalRecords, limit, offset int) *OffsetPagination {
	if limit <= 0 {
		limit = defaultLimit
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1 // always at least 1 page even when empty
	}

	currentPage := (offset / limit) + 1
	hasMore := offset+limit < totalRecords

	return &OffsetPagination{
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		CurrentPage:  currentPage,
		PerPage:      limit,
		HasMore:      hasMore,
	}
}

// CursorPagination is the metadata envelope for cursor-based (infinite scroll / feed) lists.
// Client sends ?cursor=<NextCursor> to fetch the next page.
// NextCursor = ID (UUID v7) of the last item in the current page.
// NextCursor is nil when there are no more pages.
type CursorPagination struct {
	NextCursor *string `json:"next_cursor"` // nil = last page
	HasMore    bool    `json:"has_more"`
	Limit      int     `json:"limit"`
}

// NewCursorPagination builds cursor pagination metadata.
// nextCursor should be the string ID of the last fetched item.
// Set hasMore=true and pass the last item's ID when more records exist.
// Fetch limit+1 items in the query — if count > limit, hasMore=true, trim last item.
func NewCursorPagination(nextCursor string, hasMore bool, limit int) *CursorPagination {
	var cursor *string
	if hasMore && nextCursor != "" {
		cursor = &nextCursor
	}
	return &CursorPagination{
		NextCursor: cursor,
		HasMore:    hasMore,
		Limit:      limit,
	}
}

// ─────────────────────────────────────────────────────────────────────
//  Response Envelopes
// ─────────────────────────────────────────────────────────────────────

// SuccessResponse is the top-level envelope for all successful responses.
type SuccessResponse struct {
	Data any `json:"data"`
}

// ListPayload wraps items + optional pagination metadata under data.
type ListPayload struct {
	Items      any `json:"items"`
	Pagination any `json:"pagination,omitempty"`
}

// ErrorResponse is the top-level envelope for all error responses.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail carries code, message, and optional field-level details.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// ─────────────────────────────────────────────────────────────────────
//  Render Functions
// ─────────────────────────────────────────────────────────────────────

// RenderItem sends a single object under the data envelope.
// Optional statusCode overrides 200 OK (e.g. fiber.StatusCreated).
func RenderItem(c fiber.Ctx, data any, statusCode ...int) error {
	status := fiber.StatusOK
	if len(statusCode) > 0 {
		status = statusCode[0]
	}
	return c.Status(status).JSON(SuccessResponse{Data: data})
}

// RenderList sends a collection under data.items with optional pagination metadata.
// Pass *OffsetPagination or *CursorPagination as the pagination argument.
//
// Offset example:
//
//	q := presenter.ParseOffsetQuery(c)
//	items, total := svc.List(ctx, q.Limit, q.Offset)
//	return presenter.RenderList(c, items, presenter.NewOffsetPagination(total, q.Limit, q.Offset))
//
// Cursor example:
//
//	q := presenter.ParseCursorQuery(c)
//	items := svc.ListAfter(ctx, q.Cursor, q.Limit+1)   // fetch one extra to detect hasMore
//	hasMore := len(items) > q.Limit
//	if hasMore { items = items[:q.Limit] }
//	nextCursor := ""
//	if hasMore { nextCursor = items[len(items)-1].ID.String() }
//	return presenter.RenderList(c, items, presenter.NewCursorPagination(nextCursor, hasMore, q.Limit))
func RenderList(c fiber.Ctx, items any, pagination ...any) error {
	var pg any
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

// RenderError sends a structured error response derived from AppError.
func RenderError(c fiber.Ctx, err error) error {
	if appErr, ok := err.(*custom_errors.AppError); ok {
		if appErr == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
				Error: ErrorDetail{
					Code:    "INTERNAL_SERVER_ERROR",
					Message: "An unexpected error occurred",
				},
			})
		}
		return c.Status(appErr.HTTPStatus).JSON(ErrorResponse{
			Error: ErrorDetail{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			},
		})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		Error: ErrorDetail{
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "An unexpected error occurred",
		},
	})
}
