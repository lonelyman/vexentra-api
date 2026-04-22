package projecthdl

import (
	"strings"

	"vexentra-api/internal/modules/project"
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

// ProjectHandler exposes CRUD over /api/v1/projects. All endpoints require
// a valid access token; per-action permission checks happen inside the service.
type ProjectHandler struct {
	svc      projectsvc.ProjectService
	validate *validator.Validate
	logger   logger.Logger
}

func NewProjectHandler(svc projectsvc.ProjectService, l logger.Logger) *ProjectHandler {
	if l == nil {
		l = logger.Get()
	}
	return &ProjectHandler{
		svc:      svc,
		validate: validation.New(),
		logger:   l,
	}
}

// Create — POST /api/v1/projects
func (h *ProjectHandler) Create(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}

	req := new(CreateProjectRequest)
	if err := c.Bind().Body(req); err != nil {
		return custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง")
	}
	if v := validation.Validate(h.validate, req); !v.IsValid {
		return custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", v.Errors)
	}

	clientPersonID, err := parseUUIDPtr(req.ClientPersonID)
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrValidation, "client_person_id ไม่ถูกต้อง")
	}

	p, svcErr := h.svc.Create(c.Context(), caller, projectsvc.CreateProjectInput{
		Name:             req.Name,
		Description:      req.Description,
		ClientPersonID:   clientPersonID,
		ClientNameRaw:    req.ClientNameRaw,
		ClientEmailRaw:   req.ClientEmailRaw,
		ScheduledStartAt: req.ScheduledStartAt,
		DeadlineAt:       req.DeadlineAt,
	})
	if svcErr != nil {
		return svcErr
	}
	h.logger.Success("Project created", "projectID", p.ID, "code", p.ProjectCode)
	return presenter.RenderItem(c, NewProjectResponse(p), fiber.StatusCreated)
}

// Get — GET /api/v1/projects/:id
func (h *ProjectHandler) Get(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}
	p, svcErr := h.svc.Get(c.Context(), caller, id)
	if svcErr != nil {
		return svcErr
	}
	return presenter.RenderItem(c, NewProjectResponse(p))
}

// GetByCode — GET /api/v1/projects/by-code/:code
func (h *ProjectHandler) GetByCode(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	code := strings.TrimSpace(c.Params("code"))
	if code == "" {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "กรุณาระบุรหัสโปรเจกต์")
	}
	p, svcErr := h.svc.GetByCode(c.Context(), caller, code)
	if svcErr != nil {
		return svcErr
	}
	return presenter.RenderItem(c, NewProjectResponse(p))
}

// List — GET /api/v1/projects?status=&search=&page=&limit=
func (h *ProjectHandler) List(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}

	q := presenter.ParseOffsetQuery(c)

	var statuses []project.ProjectStatus
	if raw := strings.TrimSpace(c.Query("status", "")); raw != "" {
		for _, s := range strings.Split(raw, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				statuses = append(statuses, project.ProjectStatus(s))
			}
		}
	}

	filter := project.ProjectFilter{
		Statuses: statuses,
		Search:   strings.TrimSpace(c.Query("search", "")),
	}

	items, total, svcErr := h.svc.List(c.Context(), caller, filter, project.Pagination{
		Limit:  q.Limit,
		Offset: q.Offset,
	})
	if svcErr != nil {
		return svcErr
	}

	resp := make([]ProjectResponse, len(items))
	for i, p := range items {
		resp[i] = NewProjectResponse(p)
	}
	pg := presenter.NewOffsetPagination(int(total), q.Limit, q.Offset)
	return presenter.RenderList(c, resp, pg)
}

// ListStatuses — GET /api/v1/project-statuses?active_only=true
func (h *ProjectHandler) ListStatuses(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}

	activeOnly := !strings.EqualFold(strings.TrimSpace(c.Query("active_only", "")), "false")
	items, svcErr := h.svc.ListStatuses(c.Context(), caller, activeOnly)
	if svcErr != nil {
		return svcErr
	}

	resp := make([]ProjectStatusResponse, len(items))
	for i := range items {
		resp[i] = NewProjectStatusResponse(items[i])
	}
	return presenter.RenderList(c, resp)
}

// GetFinancialPlan — GET /api/v1/projects/:id/financial-plan
func (h *ProjectHandler) GetFinancialPlan(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}
	plan, svcErr := h.svc.GetFinancialPlan(c.Context(), caller, id)
	if svcErr != nil {
		return svcErr
	}
	return presenter.RenderItem(c, NewProjectFinancialPlanResponse(plan))
}

// UpsertFinancialPlan — PUT /api/v1/projects/:id/financial-plan
func (h *ProjectHandler) UpsertFinancialPlan(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}

	req := new(UpsertProjectFinancialPlanRequest)
	if err := c.Bind().Body(req); err != nil {
		return custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง")
	}
	if v := validation.Validate(h.validate, req); !v.IsValid {
		return custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", v.Errors)
	}

	items := make([]projectsvc.ProjectPaymentInstallmentInput, 0, len(req.Installments))
	for i := range req.Installments {
		item := req.Installments[i]
		items = append(items, projectsvc.ProjectPaymentInstallmentInput{
			SortOrder:           item.SortOrder,
			Title:               item.Title,
			Amount:              item.Amount,
			PlannedDeliveryDate: item.PlannedDeliveryDate,
			PlannedReceiveDate:  item.PlannedReceiveDate,
			Note:                item.Note,
		})
	}

	plan, svcErr := h.svc.UpsertFinancialPlan(c.Context(), caller, id, projectsvc.UpsertFinancialPlanInput{
		ContractAmount:      req.ContractAmount,
		RetentionAmount:     req.RetentionAmount,
		PlannedDeliveryDate: req.PlannedDeliveryDate,
		PaymentNote:         req.PaymentNote,
		Installments:        items,
	})
	if svcErr != nil {
		return svcErr
	}
	return presenter.RenderItem(c, NewProjectFinancialPlanResponse(plan))
}

// Update — PUT /api/v1/projects/:id
func (h *ProjectHandler) Update(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}

	req := new(UpdateProjectRequest)
	if err := c.Bind().Body(req); err != nil {
		return custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง")
	}
	if v := validation.Validate(h.validate, req); !v.IsValid {
		return custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", v.Errors)
	}

	clientPersonID, err := parseUUIDPtr(req.ClientPersonID)
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrValidation, "client_person_id ไม่ถูกต้อง")
	}

	p, svcErr := h.svc.Update(c.Context(), caller, id, projectsvc.UpdateProjectInput{
		Name:             req.Name,
		Description:      req.Description,
		Status:           project.ProjectStatus(req.Status),
		ClientPersonID:   clientPersonID,
		ClientNameRaw:    req.ClientNameRaw,
		ClientEmailRaw:   req.ClientEmailRaw,
		ScheduledStartAt: req.ScheduledStartAt,
		DeadlineAt:       req.DeadlineAt,
	})
	if svcErr != nil {
		return svcErr
	}
	return presenter.RenderItem(c, NewProjectResponse(p))
}

// Close — POST /api/v1/projects/:id/close
// Closure requires a reason and is separate from Update so the closed/reason
// bijection is enforced at the boundary (DB CHECK mirrors this rule).
func (h *ProjectHandler) Close(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}

	req := new(CloseProjectRequest)
	if err := c.Bind().Body(req); err != nil {
		return custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง")
	}
	if v := validation.Validate(h.validate, req); !v.IsValid {
		return custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", v.Errors)
	}

	p, svcErr := h.svc.Close(c.Context(), caller, id, projectsvc.CloseProjectInput{
		Reason:   project.ProjectClosureReason(req.Reason),
		ClosedAt: req.ClosedAt,
	})
	if svcErr != nil {
		return svcErr
	}
	h.logger.Info("Project closed", "projectID", p.ID, "reason", req.Reason)
	return presenter.RenderItem(c, NewProjectResponse(p))
}

// Delete — DELETE /api/v1/projects/:id (soft delete)
func (h *ProjectHandler) Delete(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}
	if svcErr := h.svc.Delete(c.Context(), caller, id); svcErr != nil {
		return svcErr
	}
	return c.SendStatus(fiber.StatusNoContent)
}
