package dashboard

import (
	"time"

	"vexentra-api/internal/modules/project"

	"github.com/google/uuid"
)

type StatusCount struct {
	Status project.ProjectStatus
	Count  int64
}

type DeadlineProject struct {
	ID          uuid.UUID
	ProjectCode string
	Name        string
	Status      project.ProjectStatus
	DeadlineAt  time.Time
}

type PLSummary struct {
	Income  string // formatted NUMERIC(15,2) as string, e.g. "150000.00"
	Expense string
	Net     string
}

// Stats is the aggregate payload returned by GET /dashboard/stats.
// All figures are scoped to the caller's accessible projects (access-checked by DashboardService).
type Stats struct {
	StatusCounts      []StatusCount
	UpcomingDeadlines []DeadlineProject
	PL                PLSummary
}
