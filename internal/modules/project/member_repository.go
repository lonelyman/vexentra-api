package project

import (
	"context"

	"github.com/google/uuid"
)

type ProjectMemberRepository interface {
	Add(ctx context.Context, m *ProjectMember) error

	GetByID(ctx context.Context, id uuid.UUID) (*ProjectMember, error)

	// GetActiveByProjectAndPerson returns the active (deleted_at IS NULL) membership
	// for the (project, person) pair, or nil if none exists.
	// Used by service layer for permission checks and duplicate-add guards.
	GetActiveByProjectAndPerson(ctx context.Context, projectID, personID uuid.UUID) (*ProjectMember, error)

	// ListByProject returns active memberships of a project, ordered by created_at ASC.
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]*ProjectMember, error)

	// GetActiveLead returns the current lead of the project, or nil if none.
	GetActiveLead(ctx context.Context, projectID uuid.UUID) (*ProjectMember, error)

	// TransferLead atomically sets is_lead=true on toMemberID and is_lead=false on any
	// other active lead of the same project — satisfies the one-lead partial unique index.
	TransferLead(ctx context.Context, projectID, toMemberID uuid.UUID) error

	// Remove soft-deletes the membership. Service must refuse to remove the sole lead.
	Remove(ctx context.Context, id uuid.UUID) error

	// ListRoleMaster returns role definitions used for per-project assignments.
	ListRoleMaster(ctx context.Context, activeOnly bool) ([]*ProjectRole, error)

	// CountActiveRoleMasterByIDs validates role IDs against active master rows.
	CountActiveRoleMasterByIDs(ctx context.Context, roleIDs []uuid.UUID) (int64, error)

	// ReplaceMemberRoles soft-deletes current active assignments and inserts the new set.
	// If roleIDs is empty, this clears all roles for the member.
	ReplaceMemberRoles(ctx context.Context, memberID, assignedByUserID uuid.UUID, roleIDs []uuid.UUID, primaryRoleID *uuid.UUID) error

	// HasActiveRoleCode checks if a person has an active assignment with given role code in a project.
	HasActiveRoleCode(ctx context.Context, projectID, personID uuid.UUID, roleCode string) (bool, error)
}
