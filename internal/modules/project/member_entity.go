package project

import (
	"time"

	"github.com/google/uuid"
)

// ProjectMember represents an active or historical membership of a Person in a Project.
// Structure is flat + is_lead with optional role assignments from master data.
// Lead is transferable and only one active lead per project is allowed
// (enforced by partial unique index in migration 20260422000003).
//
// Lifecycle:
//   - Joined  → CreatedAt
//   - Left    → DeletedAt (soft delete; there is no separate left_at column)
type ProjectMember struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	PersonID  uuid.UUID

	IsLead bool
	Roles  []ProjectMemberRole

	AddedByUserID uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

// ProjectRole is master data for responsibilities assignable per project member.
type ProjectRole struct {
	ID          uuid.UUID
	Code        string
	NameTH      string
	NameEN      string
	Description *string
	SortOrder   int
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// ProjectMemberRole is an active role assignment on a project membership.
type ProjectMemberRole struct {
	AssignmentID uuid.UUID
	RoleID       uuid.UUID
	Code         string
	NameTH       string
	NameEN       string
	IsPrimary    bool
}
