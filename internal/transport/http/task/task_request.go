package taskhdl

import "time"

type CreateTaskRequest struct {
	Title            string     `json:"title"              validate:"required,min=1,max=300"`
	Description      *string    `json:"description"`
	Priority         string     `json:"priority"           validate:"omitempty,oneof=low medium high"`
	AssignedPersonID *string    `json:"assigned_person_id" validate:"omitempty,uuid"`
	DueDate          *time.Time `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title            string     `json:"title"              validate:"required,min=1,max=300"`
	Description      *string    `json:"description"`
	Status           string     `json:"status"             validate:"required,oneof=todo in_progress done cancelled"`
	Priority         string     `json:"priority"           validate:"required,oneof=low medium high"`
	AssignedPersonID *string    `json:"assigned_person_id" validate:"omitempty,uuid"`
	DueDate          *time.Time `json:"due_date"`
}
