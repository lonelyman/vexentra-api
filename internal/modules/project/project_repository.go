package project

import (
	"context"

	"github.com/google/uuid"
)

// ProjectFilter narrows List queries. All fields optional — zero value = no filter on that field.
// MemberPersonID, when set, restricts results to projects where the person is an active member
// (joined via project_members, WHERE deleted_at IS NULL). This is the primary lever for "my projects" views.
type ProjectFilter struct {
	Statuses        []ProjectStatus // OR-match; empty = all statuses
	ProjectKinds    []ProjectKind   // OR-match; empty = all kinds
	CreatedByUserID *uuid.UUID      // exact match on creator
	MemberPersonID  *uuid.UUID      // active membership filter
	ClientPersonID  *uuid.UUID      // exact match on client
	Search          string          // case-insensitive substring over name + project_code
}

// Pagination is a simple offset/limit pair. Limit must be > 0; caller enforces caps.
type Pagination struct {
	Limit  int
	Offset int
}

type ProjectRepository interface {
	Create(ctx context.Context, p *Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*Project, error)
	GetByCode(ctx context.Context, code string) (*Project, error)
	ListStatuses(ctx context.Context, activeOnly bool) ([]ProjectStatusMeta, error)
	GetFinancialPlan(ctx context.Context, projectID uuid.UUID) (*ProjectFinancialPlan, error)
	UpsertFinancialPlan(ctx context.Context, plan *ProjectFinancialPlan) error

	List(ctx context.Context, f ProjectFilter, pg Pagination) (items []*Project, total int64, err error)

	Update(ctx context.Context, p *Project) error
	Delete(ctx context.Context, id uuid.UUID) error // soft delete

	// NextCodeSeq advances and returns the next value of `project_code_seq`.
	// Service layer composes the final code string (PREFIX-YYYY-NNNN) from this.
	NextCodeSeq(ctx context.Context) (int64, error)
}
