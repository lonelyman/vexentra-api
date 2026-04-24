package filehdl

import (
	filesvc "vexentra-api/internal/modules/file/filesvc"
	"vexentra-api/internal/transport/http/presenter"
	"vexentra-api/pkg/auth"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"
	"vexentra-api/pkg/validation"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type UploadHandler struct {
	svc      filesvc.Service
	validate *validator.Validate
	logger   logger.Logger
}

func NewUploadHandler(svc filesvc.Service, l logger.Logger) *UploadHandler {
	if l == nil {
		l = logger.Get()
	}
	return &UploadHandler{
		svc:      svc,
		validate: validation.New(),
		logger:   l,
	}
}

func (h *UploadHandler) Presign(c fiber.Ctx) error {
	var req PresignUploadRequest
	if err := c.Bind().JSON(&req); err != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "รูปแบบข้อมูลไม่ถูกต้อง"))
	}
	if v := validation.Validate(h.validate, &req); !v.IsValid {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", v.Errors))
	}

	caller, err := auth.GetCaller(c)
	if err != nil {
		return presenter.RenderError(c, err)
	}

	var targetPersonID *uuid.UUID
	if req.TargetPersonID != "" {
		parsed, parseErr := uuid.Parse(req.TargetPersonID)
		if parseErr != nil {
			return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "target person id ไม่ถูกต้อง"))
		}
		targetPersonID = &parsed
	}

	result, svcErr := h.svc.PresignProfileImage(c.Context(), caller, req.Filename, req.MIMEType, req.SizeBytes, targetPersonID)
	if svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}
	return presenter.RenderItem(c, result, fiber.StatusCreated)
}

func (h *UploadHandler) Complete(c fiber.Ctx) error {
	var req CompleteUploadRequest
	if err := c.Bind().JSON(&req); err != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "รูปแบบข้อมูลไม่ถูกต้อง"))
	}
	if v := validation.Validate(h.validate, &req); !v.IsValid {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", v.Errors))
	}

	sessionID, err := uuid.Parse(req.UploadSessionID)
	if err != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "session id ไม่ถูกต้อง"))
	}

	caller, cerr := auth.GetCaller(c)
	if cerr != nil {
		return presenter.RenderError(c, cerr)
	}
	result, svcErr := h.svc.CompleteProfileImage(c.Context(), caller, sessionID)
	if svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}
	return presenter.RenderItem(c, result)
}

func (h *UploadHandler) GetFileURL(c fiber.Ctx) error {
	fileID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "file id ไม่ถูกต้อง"))
	}
	caller, cerr := auth.GetCaller(c)
	if cerr != nil {
		return presenter.RenderError(c, cerr)
	}
	url, svcErr := h.svc.GetFileURL(c.Context(), caller, fileID)
	if svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}
	return presenter.RenderItem(c, fiber.Map{"url": url})
}

func (h *UploadHandler) Delete(c fiber.Ctx) error {
	fileID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "file id ไม่ถูกต้อง"))
	}
	caller, cerr := auth.GetCaller(c)
	if cerr != nil {
		return presenter.RenderError(c, cerr)
	}
	if svcErr := h.svc.DeleteFile(c.Context(), caller, fileID); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}
	return presenter.RenderItem(c, fiber.Map{"message": "ลบไฟล์สำเร็จ"})
}
