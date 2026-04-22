package projecthdl

import (
	"vexentra-api/internal/modules/project/projectsvc"
	"vexentra-api/internal/transport/http/presenter"
	"vexentra-api/pkg/auth"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"
	"vexentra-api/pkg/validation"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type MemberHandler struct {
	svc      projectsvc.MemberService
	validate *validator.Validate
	logger   logger.Logger
}

func NewMemberHandler(svc projectsvc.MemberService, l logger.Logger) *MemberHandler {
	if l == nil {
		l = logger.Get()
	}
	return &MemberHandler{svc: svc, validate: validation.New(), logger: l}
}

// Add — POST /api/v1/projects/:id/members
func (h *MemberHandler) Add(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}

	req := new(AddMemberRequest)
	if err := c.Bind().Body(req); err != nil {
		return custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง")
	}
	if v := validation.Validate(h.validate, req); !v.IsValid {
		return custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", v.Errors)
	}
	personID, err := uuid.Parse(req.PersonID)
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrValidation, "person_id ไม่ถูกต้อง")
	}

	m, svcErr := h.svc.Add(c.Context(), caller, projectID, personID)
	if svcErr != nil {
		return svcErr
	}
	h.logger.Info("Member added", "projectID", projectID, "personID", personID)
	return presenter.RenderItem(c, NewMemberResponse(m), fiber.StatusCreated)
}

// List — GET /api/v1/projects/:id/members
func (h *MemberHandler) List(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}
	items, svcErr := h.svc.List(c.Context(), caller, projectID)
	if svcErr != nil {
		return svcErr
	}
	resp := make([]MemberResponse, len(items))
	for i, m := range items {
		resp[i] = NewMemberResponse(m)
	}
	return presenter.RenderList(c, resp)
}

// Remove — DELETE /api/v1/projects/:id/members/:memberID
func (h *MemberHandler) Remove(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}
	memberID, err := uuid.Parse(c.Params("memberID"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "member id ไม่ถูกต้อง")
	}
	if svcErr := h.svc.Remove(c.Context(), caller, projectID, memberID); svcErr != nil {
		return svcErr
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// TransferLead — POST /api/v1/projects/:id/transfer-lead
func (h *MemberHandler) TransferLead(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}

	req := new(TransferLeadRequest)
	if err := c.Bind().Body(req); err != nil {
		return custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง")
	}
	if v := validation.Validate(h.validate, req); !v.IsValid {
		return custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", v.Errors)
	}
	memberID, err := uuid.Parse(req.MemberID)
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrValidation, "member_id ไม่ถูกต้อง")
	}

	if svcErr := h.svc.TransferLead(c.Context(), caller, projectID, memberID); svcErr != nil {
		return svcErr
	}
	h.logger.Info("Project lead transferred", "projectID", projectID, "toMemberID", memberID)
	return presenter.RenderItem(c, fiber.Map{"message": "โอนสิทธิ์หัวหน้าทีมสำเร็จ"})
}
