package task

import (
	"context"

	"github.com/google/uuid"
)

type TaskRepository interface {
	Create(ctx context.Context, t *Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*Task, error)

	// ListByProject returns tasks for a project ordered by priority DESC, created_at ASC.
	ListByProject(ctx context.Context, projectID uuid.UUID, f TaskFilter, pg Pagination) (items []*Task, total int64, err error)

	Update(ctx context.Context, t *Task) error
	Delete(ctx context.Context, id uuid.UUID) error // soft delete
}
