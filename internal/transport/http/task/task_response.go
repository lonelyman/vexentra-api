package taskhdl

import (
	"vexentra-api/internal/modules/task"
)

type TaskResponse struct {
	ID               string  `json:"id"`
	ProjectID        string  `json:"project_id"`
	Title            string  `json:"title"`
	Description      *string `json:"description"`
	Status           string  `json:"status"`
	Priority         string  `json:"priority"`
	AssignedPersonID *string `json:"assigned_person_id"`
	DueDate          *string `json:"due_date"` // "YYYY-MM-DD" or null
	CreatedByUserID  string  `json:"created_by_user_id"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

func NewTaskResponse(t *task.Task) TaskResponse {
	var dueDate *string
	if t.DueDate != nil {
		s := t.DueDate.Format("2006-01-02")
		dueDate = &s
	}
	var assignedPersonID *string
	if t.AssignedPersonID != nil {
		s := t.AssignedPersonID.String()
		assignedPersonID = &s
	}
	return TaskResponse{
		ID:               t.ID.String(),
		ProjectID:        t.ProjectID.String(),
		Title:            t.Title,
		Description:      t.Description,
		Status:           string(t.Status),
		Priority:         string(t.Priority),
		AssignedPersonID: assignedPersonID,
		DueDate:          dueDate,
		CreatedByUserID:  t.CreatedByUserID.String(),
		CreatedAt:        t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
