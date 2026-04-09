package socialplatformhdl

import (
	"vexentra-api/internal/modules/socialplatform"
	"vexentra-api/internal/modules/socialplatform/platformsvc"
	"vexentra-api/internal/transport/http/presenter"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type SocialPlatformHandler struct {
	svc    platformsvc.SocialPlatformService
	logger logger.Logger
}

func NewSocialPlatformHandler(svc platformsvc.SocialPlatformService, l logger.Logger) *SocialPlatformHandler {
	if l == nil {
		l = logger.Get()
	}
	return &SocialPlatformHandler{svc: svc, logger: l}
}

// List — GET /api/v1/social-platforms
func (h *SocialPlatformHandler) List(c fiber.Ctx) error {
	platforms, err := h.svc.List(c.Context())
	if err != nil {
		return presenter.RenderError(c, err)
	}
	resp := make([]SocialPlatformResponse, len(platforms))
	for i, p := range platforms {
		resp[i] = toResponse(p)
	}
	return presenter.RenderList(c, resp)
}

// Create — POST /api/v1/social-platforms
func (h *SocialPlatformHandler) Create(c fiber.Ctx) error {
	var req CreateSocialPlatformRequest
	if bindErr := c.Bind().JSON(&req); bindErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "รูปแบบข้อมูลไม่ถูกต้อง"))
	}

	p := &socialplatform.SocialPlatform{
		Key:       req.Key,
		Name:      req.Name,
		IconURL:   req.IconURL,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
	}

	created, svcErr := h.svc.Create(c.Context(), p)
	if svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toResponse(created), fiber.StatusCreated)
}

// Update — PUT /api/v1/social-platforms/:id
func (h *SocialPlatformHandler) Update(c fiber.Ctx) error {
	id, parseErr := uuid.Parse(c.Params("id"))
	if parseErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "platform ID ไม่ถูกต้อง"))
	}

	var req UpdateSocialPlatformRequest
	if bindErr := c.Bind().JSON(&req); bindErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "รูปแบบข้อมูลไม่ถูกต้อง"))
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	p := &socialplatform.SocialPlatform{
		Name:      req.Name,
		IconURL:   req.IconURL,
		SortOrder: req.SortOrder,
		IsActive:  isActive,
	}

	updated, svcErr := h.svc.Update(c.Context(), id, p)
	if svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toResponse(updated))
}

// Delete — DELETE /api/v1/social-platforms/:id
func (h *SocialPlatformHandler) Delete(c fiber.Ctx) error {
	id, parseErr := uuid.Parse(c.Params("id"))
	if parseErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "platform ID ไม่ถูกต้อง"))
	}

	if svcErr := h.svc.Delete(c.Context(), id); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
