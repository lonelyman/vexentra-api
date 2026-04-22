package task

import (
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

type Task struct {
	ID               uuid.UUID
	ProjectID        uuid.UUID
	Title            string
	Description      *string
	Status           TaskStatus
	Priority         TaskPriority
	AssignedPersonID *uuid.UUID
	DueDate          *time.Time

	CreatedByUserID uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type TaskFilter struct {
	Statuses   []TaskStatus
	AssignedTo *uuid.UUID // person_id
	Search     string
}

type Pagination struct {
	Limit  int
	Offset int
}
