package pgtask

import (
	"context"
	"time"

	"vexentra-api/internal/modules/task"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type taskModel struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey"`
	ProjectID        uuid.UUID      `gorm:"type:uuid;column:project_id;not null"`
	Title            string         `gorm:"column:title;not null"`
	Description      *string        `gorm:"column:description"`
	Status           string         `gorm:"column:status;not null;default:'todo'"`
	Priority         string         `gorm:"column:priority;not null;default:'medium'"`
	AssignedPersonID *uuid.UUID     `gorm:"type:uuid;column:assigned_person_id"`
	DueDate          *time.Time     `gorm:"column:due_date"`
	CreatedByUserID  uuid.UUID      `gorm:"type:uuid;column:created_by_user_id;not null"`
	CreatedAt        time.Time      `gorm:"column:created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (taskModel) TableName() string { return "tasks" }

func toModel(t *task.Task) *taskModel {
	return &taskModel{
		ID:               t.ID,
		ProjectID:        t.ProjectID,
		Title:            t.Title,
		Description:      t.Description,
		Status:           string(t.Status),
		Priority:         string(t.Priority),
		AssignedPersonID: t.AssignedPersonID,
		DueDate:          t.DueDate,
		CreatedByUserID:  t.CreatedByUserID,
	}
}

func toEntity(m *taskModel) *task.Task {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}
	return &task.Task{
		ID:               m.ID,
		ProjectID:        m.ProjectID,
		Title:            m.Title,
		Description:      m.Description,
		Status:           task.TaskStatus(m.Status),
		Priority:         task.TaskPriority(m.Priority),
		AssignedPersonID: m.AssignedPersonID,
		DueDate:          m.DueDate,
		CreatedByUserID:  m.CreatedByUserID,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
		DeletedAt:        deletedAt,
	}
}

// ─── Repository ───────────────────────────────────────────────────────────────

type taskRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

func NewTaskRepository(db *gorm.DB, l logger.Logger) task.TaskRepository {
	if l == nil {
		l = logger.Get()
	}
	return &taskRepository{db: db, logger: l}
}

func (r *taskRepository) Create(ctx context.Context, t *task.Task) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	t.ID = id

	m := toModel(t)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		r.logger.Error("DB_CREATE_TASK_ERROR", err)
		return err
	}
	t.CreatedAt = m.CreatedAt
	t.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *taskRepository) GetByID(ctx context.Context, id uuid.UUID) (*task.Task, error) {
	var m taskModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("DB_GET_TASK_ERROR", err)
		return nil, err
	}
	return toEntity(&m), nil
}

func (r *taskRepository) ListByProject(ctx context.Context, projectID uuid.UUID, f task.TaskFilter, pg task.Pagination) ([]*task.Task, int64, error) {
	q := r.db.WithContext(ctx).Model(&taskModel{}).Where("project_id = ?", projectID)

	if len(f.Statuses) > 0 {
		ss := make([]string, len(f.Statuses))
		for i, s := range f.Statuses {
			ss[i] = string(s)
		}
		q = q.Where("status IN ?", ss)
	}
	if f.AssignedTo != nil {
		q = q.Where("assigned_person_id = ?", f.AssignedTo)
	}
	if f.Search != "" {
		q = q.Where("title ILIKE ?", "%"+f.Search+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []taskModel
	orderExpr := `CASE priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END, created_at ASC`
	if err := q.Order(orderExpr).Limit(pg.Limit).Offset(pg.Offset).Find(&models).Error; err != nil {
		r.logger.Error("DB_LIST_TASKS_ERROR", err)
		return nil, 0, err
	}

	items := make([]*task.Task, len(models))
	for i, m := range models {
		m := m
		items[i] = toEntity(&m)
	}
	return items, total, nil
}

func (r *taskRepository) Update(ctx context.Context, t *task.Task) error {
	result := r.db.WithContext(ctx).
		Model(&taskModel{}).
		Where("id = ?", t.ID).
		Updates(map[string]any{
			"title":              t.Title,
			"description":        t.Description,
			"status":             string(t.Status),
			"priority":           string(t.Priority),
			"assigned_person_id": t.AssignedPersonID,
			"due_date":           t.DueDate,
		})
	if result.Error != nil {
		r.logger.Error("DB_UPDATE_TASK_ERROR", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบงานนี้")
	}
	return nil
}

func (r *taskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&taskModel{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("DB_DELETE_TASK_ERROR", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบงานนี้")
	}
	return nil
}
