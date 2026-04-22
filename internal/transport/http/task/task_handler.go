package taskhdl

import (
	"strings"

	"vexentra-api/internal/modules/task"
	"vexentra-api/internal/modules/task/tasksvc"
	"vexentra-api/internal/transport/http/presenter"
	"vexentra-api/pkg/auth"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"
	"vexentra-api/pkg/validation"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type TaskHandler struct {
	svc      tasksvc.TaskService
	validate *validator.Validate
	logger   logger.Logger
}

func NewTaskHandler(svc tasksvc.TaskService, l logger.Logger) *TaskHandler {
	if l == nil {
		l = logger.Get()
	}
	return &TaskHandler{svc: svc, validate: validation.New(), logger: l}
}

// Create — POST /api/v1/projects/:id/tasks
func (h *TaskHandler) Create(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}

	req := new(CreateTaskRequest)
	if err := c.Bind().Body(req); err != nil {
		return custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง")
	}
	if v := validation.Validate(h.validate, req); !v.IsValid {
		return custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", v.Errors)
	}

	assignedPersonID, err := parseUUIDPtr(req.AssignedPersonID)
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrValidation, "assigned_person_id ไม่ถูกต้อง")
	}

	priority := task.TaskPriority(req.Priority)
	if req.Priority == "" {
		priority = task.TaskPriorityMedium
	}

	t, svcErr := h.svc.Create(c.Context(), caller, projectID, tasksvc.CreateTaskInput{
		Title:            req.Title,
		Description:      req.Description,
		Priority:         priority,
		AssignedPersonID: assignedPersonID,
		DueDate:          req.DueDate,
	})
	if svcErr != nil {
		return svcErr
	}
	return presenter.RenderItem(c, NewTaskResponse(t), fiber.StatusCreated)
}

// List — GET /api/v1/projects/:id/tasks?status=&search=&page=&limit=
func (h *TaskHandler) List(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}

	q := presenter.ParseOffsetQuery(c)

	var statuses []task.TaskStatus
	if raw := strings.TrimSpace(c.Query("status", "")); raw != "" {
		for _, s := range strings.Split(raw, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				statuses = append(statuses, task.TaskStatus(s))
			}
		}
	}

	f := task.TaskFilter{
		Statuses: statuses,
		Search:   strings.TrimSpace(c.Query("search", "")),
	}

	items, total, svcErr := h.svc.List(c.Context(), caller, projectID, f, task.Pagination{
		Limit:  q.Limit,
		Offset: q.Offset,
	})
	if svcErr != nil {
		return svcErr
	}

	resp := make([]TaskResponse, len(items))
	for i, t := range items {
		resp[i] = NewTaskResponse(t)
	}
	pg := presenter.NewOffsetPagination(int(total), q.Limit, q.Offset)
	return presenter.RenderList(c, resp, pg)
}

// Get — GET /api/v1/projects/:id/tasks/:taskID
func (h *TaskHandler) Get(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}
	taskID, err := uuid.Parse(c.Params("taskID"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "task id ไม่ถูกต้อง")
	}

	t, svcErr := h.svc.Get(c.Context(), caller, projectID, taskID)
	if svcErr != nil {
		return svcErr
	}
	return presenter.RenderItem(c, NewTaskResponse(t))
}

// Update — PUT /api/v1/projects/:id/tasks/:taskID
func (h *TaskHandler) Update(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}
	taskID, err := uuid.Parse(c.Params("taskID"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "task id ไม่ถูกต้อง")
	}

	req := new(UpdateTaskRequest)
	if err := c.Bind().Body(req); err != nil {
		return custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง")
	}
	if v := validation.Validate(h.validate, req); !v.IsValid {
		return custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", v.Errors)
	}

	assignedPersonID, err := parseUUIDPtr(req.AssignedPersonID)
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrValidation, "assigned_person_id ไม่ถูกต้อง")
	}

	t, svcErr := h.svc.Update(c.Context(), caller, projectID, taskID, tasksvc.UpdateTaskInput{
		Title:            req.Title,
		Description:      req.Description,
		Status:           task.TaskStatus(req.Status),
		Priority:         task.TaskPriority(req.Priority),
		AssignedPersonID: assignedPersonID,
		DueDate:          req.DueDate,
	})
	if svcErr != nil {
		return svcErr
	}
	return presenter.RenderItem(c, NewTaskResponse(t))
}

// Delete — DELETE /api/v1/projects/:id/tasks/:taskID
func (h *TaskHandler) Delete(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}
	taskID, err := uuid.Parse(c.Params("taskID"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "task id ไม่ถูกต้อง")
	}

	if svcErr := h.svc.Delete(c.Context(), caller, projectID, taskID); svcErr != nil {
		return svcErr
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func parseUUIDPtr(s *string) (*uuid.UUID, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
