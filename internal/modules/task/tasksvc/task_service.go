package tasksvc

import (
	"context"
	"strings"
	"time"

	"vexentra-api/internal/modules/project"
	"vexentra-api/internal/modules/project/projectsvc"
	"vexentra-api/internal/modules/task"
	"vexentra-api/internal/modules/user"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/google/uuid"
)

type CreateTaskInput struct {
	Title            string
	Description      *string
	Priority         task.TaskPriority
	AssignedPersonID *uuid.UUID
	DueDate          *time.Time
}

type UpdateTaskInput struct {
	Title            string
	Description      *string
	Status           task.TaskStatus
	Priority         task.TaskPriority
	AssignedPersonID *uuid.UUID
	DueDate          *time.Time
}

type TaskService interface {
	Create(ctx context.Context, caller user.Caller, projectID uuid.UUID, in CreateTaskInput) (*task.Task, error)
	Get(ctx context.Context, caller user.Caller, projectID, taskID uuid.UUID) (*task.Task, error)
	List(ctx context.Context, caller user.Caller, projectID uuid.UUID, f task.TaskFilter, pg task.Pagination) ([]*task.Task, int64, error)
	Update(ctx context.Context, caller user.Caller, projectID, taskID uuid.UUID, in UpdateTaskInput) (*task.Task, error)
	Delete(ctx context.Context, caller user.Caller, projectID, taskID uuid.UUID) error
}

type taskService struct {
	projectSvc projectsvc.ProjectService
	memberRepo project.ProjectMemberRepository
	taskRepo   task.TaskRepository
	logger     logger.Logger
}

const projectCoordinatorRoleCode = "coordinator"

func New(
	projectSvc projectsvc.ProjectService,
	memberRepo project.ProjectMemberRepository,
	taskRepo task.TaskRepository,
	l logger.Logger,
) TaskService {
	if l == nil {
		l = logger.Get()
	}
	return &taskService{
		projectSvc: projectSvc,
		memberRepo: memberRepo,
		taskRepo:   taskRepo,
		logger:     l,
	}
}

// Create adds a task to the project.
// Write access is restricted to staff, lead, or coordinator.
func (s *taskService) Create(ctx context.Context, caller user.Caller, projectID uuid.UUID, in CreateTaskInput) (*task.Task, error) {
	if err := s.requireWriteAccess(ctx, caller, projectID); err != nil {
		return nil, err
	}
	if err := s.validateAssignee(ctx, projectID, in.AssignedPersonID); err != nil {
		return nil, err
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "ชื่องานห้ามว่าง")
	}
	priority := in.Priority
	if !isValidPriority(priority) {
		priority = task.TaskPriorityMedium
	}

	t := &task.Task{
		ProjectID:        projectID,
		Title:            title,
		Description:      in.Description,
		Status:           task.TaskStatusTodo,
		Priority:         priority,
		AssignedPersonID: in.AssignedPersonID,
		DueDate:          in.DueDate,
		CreatedByUserID:  caller.UserID,
	}
	if err := s.taskRepo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// Get returns a task if the caller can access the project.
func (s *taskService) Get(ctx context.Context, caller user.Caller, projectID, taskID uuid.UUID) (*task.Task, error) {
	if _, err := s.projectSvc.CanAccessProject(ctx, caller, projectID); err != nil {
		return nil, err
	}
	t, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if t == nil || t.ProjectID != projectID {
		return nil, custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบงานนี้")
	}
	return t, nil
}

// List returns paginated tasks for the project. Any project member can view.
func (s *taskService) List(ctx context.Context, caller user.Caller, projectID uuid.UUID, f task.TaskFilter, pg task.Pagination) ([]*task.Task, int64, error) {
	if _, err := s.projectSvc.CanAccessProject(ctx, caller, projectID); err != nil {
		return nil, 0, err
	}
	if pg.Limit <= 0 || pg.Limit > 200 {
		pg.Limit = 100
	}
	if pg.Offset < 0 {
		pg.Offset = 0
	}
	return s.taskRepo.ListByProject(ctx, projectID, f, pg)
}

// Update task details.
// Write access is restricted to staff, lead, or coordinator.
func (s *taskService) Update(ctx context.Context, caller user.Caller, projectID, taskID uuid.UUID, in UpdateTaskInput) (*task.Task, error) {
	if err := s.requireWriteAccess(ctx, caller, projectID); err != nil {
		return nil, err
	}
	if err := s.validateAssignee(ctx, projectID, in.AssignedPersonID); err != nil {
		return nil, err
	}

	t, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if t == nil || t.ProjectID != projectID {
		return nil, custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบงานนี้")
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "ชื่องานห้ามว่าง")
	}
	if !isValidStatus(in.Status) {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "สถานะงานไม่ถูกต้อง")
	}
	if !isValidPriority(in.Priority) {
		return nil, custom_errors.New(400, custom_errors.ErrValidation, "ระดับความสำคัญไม่ถูกต้อง")
	}

	t.Title = title
	t.Description = in.Description
	t.Status = in.Status
	t.Priority = in.Priority
	t.AssignedPersonID = in.AssignedPersonID
	t.DueDate = in.DueDate

	if err := s.taskRepo.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// Delete soft-deletes a task.
// Write access is restricted to staff, lead, or coordinator.
func (s *taskService) Delete(ctx context.Context, caller user.Caller, projectID, taskID uuid.UUID) error {
	if err := s.requireWriteAccess(ctx, caller, projectID); err != nil {
		return err
	}
	t, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if t == nil || t.ProjectID != projectID {
		return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบงานนี้")
	}
	return s.taskRepo.Delete(ctx, taskID)
}

func (s *taskService) requireWriteAccess(ctx context.Context, caller user.Caller, projectID uuid.UUID) error {
	if _, err := s.projectSvc.CanAccessProject(ctx, caller, projectID); err != nil {
		return err
	}
	if caller.IsStaff() {
		return nil
	}
	lead, err := s.memberRepo.GetActiveLead(ctx, projectID)
	if err != nil {
		return err
	}
	if lead != nil && lead.PersonID == caller.PersonID {
		return nil
	}
	isCoordinator, err := s.memberRepo.HasActiveRoleCode(ctx, projectID, caller.PersonID, projectCoordinatorRoleCode)
	if err != nil {
		return err
	}
	if isCoordinator {
		return nil
	}
	return custom_errors.New(403, custom_errors.ErrForbidden, "member ทั่วไปเป็น read-only (แก้ไข task ได้เฉพาะหัวหน้าทีม, coordinator หรือผู้ดูแลระบบ)")
}

func isValidStatus(s task.TaskStatus) bool {
	switch s {
	case task.TaskStatusTodo, task.TaskStatusInProgress, task.TaskStatusDone, task.TaskStatusCancelled:
		return true
	}
	return false
}

func isValidPriority(p task.TaskPriority) bool {
	switch p {
	case task.TaskPriorityLow, task.TaskPriorityMedium, task.TaskPriorityHigh:
		return true
	}
	return false
}

func (s *taskService) validateAssignee(ctx context.Context, projectID uuid.UUID, assignedPersonID *uuid.UUID) error {
	if assignedPersonID == nil {
		return nil
	}
	m, err := s.memberRepo.GetActiveByProjectAndPerson(ctx, projectID, *assignedPersonID)
	if err != nil {
		return err
	}
	if m == nil {
		return custom_errors.New(400, custom_errors.ErrValidation, "assigned_person_id ต้องเป็นสมาชิกโปรเจกต์")
	}
	return nil
}
