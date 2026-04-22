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
}
