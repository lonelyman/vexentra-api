package project

import (
	"time"

	"github.com/google/uuid"
)

// ProjectMember represents an active or historical membership of a Person in a Project.
// Structure is flat + is_lead (no per-member role column) — project creator becomes
// the first lead; lead is transferable but only one active lead per project is allowed
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

	AddedByUserID uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}
