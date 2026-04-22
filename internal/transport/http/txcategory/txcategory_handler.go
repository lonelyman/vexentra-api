package txcategoryhdl

import (
	"strings"

	"vexentra-api/internal/modules/txcategory"
	"vexentra-api/internal/modules/txcategory/txcategorysvc"
	projecthdl "vexentra-api/internal/transport/http/project"
	"vexentra-api/internal/transport/http/presenter"
	"vexentra-api/pkg/auth"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"
	"vexentra-api/pkg/validation"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type CreateCategoryRequest struct {
	Code      string  `json:"code"       validate:"required"`
	Name      string  `json:"name"       validate:"required"`
	Type      string  `json:"type"       validate:"required,oneof=income expense"`
	IconKey   *string `json:"icon_key"`
	IsActive  bool    `json:"is_active"`
	SortOrder int     `json:"sort_order"`
}

type UpdateCategoryRequest struct {
	Name      string  `json:"name"       validate:"required"`
	IconKey   *string `json:"icon_key"`
	IsActive  bool    `json:"is_active"`
	SortOrder int     `json:"sort_order"`
}

type CategoryHandler struct {
	svc      txcategorysvc.TransactionCategoryService
	validate *validator.Validate
	logger   logger.Logger
}

func NewCategoryHandler(svc txcategorysvc.TransactionCategoryService, l logger.Logger) *CategoryHandler {
	if l == nil {
		l = logger.Get()
	}
	return &CategoryHandler{svc: svc, validate: validation.New(), logger: l}
}

// List — GET /api/v1/tx-categories?type=&active_only=&include_deleted=
func (h *CategoryHandler) List(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}

	filter := txcategory.TransactionCategoryFilter{
		ActiveOnly:     strings.EqualFold(c.Query("active_only", ""), "true"),
		IncludeDeleted: strings.EqualFold(c.Query("include_deleted", ""), "true"),
	}
	if t := strings.TrimSpace(c.Query("type", "")); t != "" {
		tt := txcategory.TransactionType(t)
		filter.Type = &tt
	}

	items, svcErr := h.svc.List(c.Context(), caller, filter)
	if svcErr != nil {
		return svcErr
	}
	resp := make([]projecthdl.TransactionCategoryResponse, len(items))
	for i, it := range items {
		resp[i] = projecthdl.NewTransactionCategoryResponse(it)
	}
	return presenter.RenderList(c, resp)
}

// Get — GET /api/v1/tx-categories/:id
func (h *CategoryHandler) Get(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "category id ไม่ถูกต้อง")
	}
	cat, svcErr := h.svc.Get(c.Context(), caller, id)
	if svcErr != nil {
		return svcErr
	}
	return presenter.RenderItem(c, projecthdl.NewTransactionCategoryResponse(cat))
}

// Create — POST /api/v1/tx-categories (admin only; enforced in service)
func (h *CategoryHandler) Create(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	req := new(CreateCategoryRequest)
	if err := c.Bind().Body(req); err != nil {
		return custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง")
	}
	if v := validation.Validate(h.validate, req); !v.IsValid {
		return custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", v.Errors)
	}

	cat, svcErr := h.svc.Create(c.Context(), caller, txcategorysvc.CreateCategoryInput{
		Code:      req.Code,
		Name:      req.Name,
		Type:      txcategory.TransactionType(req.Type),
		IconKey:   req.IconKey,
		IsActive:  req.IsActive,
		SortOrder: req.SortOrder,
	})
	if svcErr != nil {
		return svcErr
	}
	h.logger.Success("TxCategory created", "id", cat.ID, "code", cat.Code)
	return presenter.RenderItem(c, projecthdl.NewTransactionCategoryResponse(cat), fiber.StatusCreated)
}

// Update — PUT /api/v1/tx-categories/:id
func (h *CategoryHandler) Update(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "category id ไม่ถูกต้อง")
	}
	req := new(UpdateCategoryRequest)
	if err := c.Bind().Body(req); err != nil {
		return custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง")
	}
	if v := validation.Validate(h.validate, req); !v.IsValid {
		return custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", v.Errors)
	}
	cat, svcErr := h.svc.Update(c.Context(), caller, id, txcategorysvc.UpdateCategoryInput{
		Name:      req.Name,
		IconKey:   req.IconKey,
		IsActive:  req.IsActive,
		SortOrder: req.SortOrder,
	})
	if svcErr != nil {
		return svcErr
	}
	return presenter.RenderItem(c, projecthdl.NewTransactionCategoryResponse(cat))
}

// Delete — DELETE /api/v1/tx-categories/:id
func (h *CategoryHandler) Delete(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "category id ไม่ถูกต้อง")
	}
	if svcErr := h.svc.Delete(c.Context(), caller, id); svcErr != nil {
		return svcErr
	}
	return c.SendStatus(fiber.StatusNoContent)
}
