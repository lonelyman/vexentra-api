package dashboardsvc

import (
	"context"
	"fmt"
	"time"

	"vexentra-api/internal/modules/dashboard"
	"vexentra-api/internal/modules/project"
	"vexentra-api/internal/modules/user"
	"vexentra-api/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DashboardService runs cross-domain read-only aggregate queries. It receives *gorm.DB
// directly because its queries span projects + transactions + members — no single repo owns them.
type DashboardService interface {
	GetStats(ctx context.Context, caller user.Caller) (*dashboard.Stats, error)
}

type dashboardService struct {
	db     *gorm.DB
	logger logger.Logger
}

func New(db *gorm.DB, l logger.Logger) DashboardService {
	if l == nil {
		l = logger.Get()
	}
	return &dashboardService{db: db, logger: l}
}

func (s *dashboardService) GetStats(ctx context.Context, caller user.Caller) (*dashboard.Stats, error) {
	db := s.db.WithContext(ctx)
	nonStaff := !caller.IsStaff()
	pid := caller.PersonID

	statusCounts, err := s.queryStatusCounts(db, nonStaff, pid)
	if err != nil {
		return nil, err
	}

	deadlines, err := s.queryDeadlines(db, nonStaff, pid)
	if err != nil {
		return nil, err
	}

	pl, err := s.queryPL(db, nonStaff, pid)
	if err != nil {
		return nil, err
	}

	return &dashboard.Stats{
		StatusCounts:      statusCounts,
		UpcomingDeadlines: deadlines,
		PL:                *pl,
	}, nil
}

// ─── Status counts (non-closed projects only) ─────────────────────────────────

func (s *dashboardService) queryStatusCounts(db *gorm.DB, nonStaff bool, pid uuid.UUID) ([]dashboard.StatusCount, error) {
	type row struct {
		Status string
		Count  int64
	}
	var rows []row

	var err error
	if nonStaff {
		err = db.Raw(`
			SELECT p.status, COUNT(*) AS count
			FROM projects p
			JOIN project_members pm
				ON pm.project_id = p.id AND pm.deleted_at IS NULL AND pm.person_id = ?
			WHERE p.deleted_at IS NULL AND p.status != 'closed'
			GROUP BY p.status
		`, pid).Scan(&rows).Error
	} else {
		err = db.Raw(`
			SELECT status, COUNT(*) AS count
			FROM projects
			WHERE deleted_at IS NULL AND status != 'closed'
			GROUP BY status
		`).Scan(&rows).Error
	}
	if err != nil {
		return nil, err
	}

	counts := make([]dashboard.StatusCount, len(rows))
	for i, r := range rows {
		counts[i] = dashboard.StatusCount{
			Status: project.ProjectStatus(r.Status),
			Count:  r.Count,
		}
	}
	return counts, nil
}

// ─── Upcoming deadlines (next 30 days, non-closed) ────────────────────────────

func (s *dashboardService) queryDeadlines(db *gorm.DB, nonStaff bool, pid uuid.UUID) ([]dashboard.DeadlineProject, error) {
	type row struct {
		ID          uuid.UUID
		ProjectCode string
		Name        string
		Status      string
		DeadlineAt  time.Time
	}
	var rows []row

	var err error
	if nonStaff {
		err = db.Raw(`
			SELECT p.id, p.project_code, p.name, p.status, p.deadline_at
			FROM projects p
			JOIN project_members pm
				ON pm.project_id = p.id AND pm.deleted_at IS NULL AND pm.person_id = ?
			WHERE p.deleted_at IS NULL
			  AND p.status != 'closed'
			  AND p.deadline_at IS NOT NULL
			  AND p.deadline_at BETWEEN NOW() AND NOW() + INTERVAL '30 days'
			ORDER BY p.deadline_at ASC
			LIMIT 10
		`, pid).Scan(&rows).Error
	} else {
		err = db.Raw(`
			SELECT id, project_code, name, status, deadline_at
			FROM projects
			WHERE deleted_at IS NULL
			  AND status != 'closed'
			  AND deadline_at IS NOT NULL
			  AND deadline_at BETWEEN NOW() AND NOW() + INTERVAL '30 days'
			ORDER BY deadline_at ASC
			LIMIT 10
		`).Scan(&rows).Error
	}
	if err != nil {
		return nil, err
	}

	result := make([]dashboard.DeadlineProject, len(rows))
	for i, r := range rows {
		result[i] = dashboard.DeadlineProject{
			ID:          r.ID,
			ProjectCode: r.ProjectCode,
			Name:        r.Name,
			Status:      project.ProjectStatus(r.Status),
			DeadlineAt:  r.DeadlineAt,
		}
	}
	return result, nil
}

// ─── Global P&L (non-closed projects only) ────────────────────────────────────

func (s *dashboardService) queryPL(db *gorm.DB, nonStaff bool, pid uuid.UUID) (*dashboard.PLSummary, error) {
	type row struct {
		Income  float64
		Expense float64
	}
	var r row

	var err error
	if nonStaff {
		err = db.Raw(`
			SELECT
				COALESCE(SUM(CASE WHEN tc.type = 'income'  THEN pt.amount ELSE 0 END), 0) AS income,
				COALESCE(SUM(CASE WHEN tc.type = 'expense' THEN pt.amount ELSE 0 END), 0) AS expense
			FROM project_transactions pt
			JOIN transaction_categories tc ON tc.id = pt.category_id AND tc.deleted_at IS NULL
			JOIN projects p ON p.id = pt.project_id AND p.deleted_at IS NULL AND p.status != 'closed'
			JOIN project_members pm
				ON pm.project_id = p.id AND pm.deleted_at IS NULL AND pm.person_id = ?
			WHERE pt.deleted_at IS NULL
		`, pid).Scan(&r).Error
	} else {
		err = db.Raw(`
			SELECT
				COALESCE(SUM(CASE WHEN tc.type = 'income'  THEN pt.amount ELSE 0 END), 0) AS income,
				COALESCE(SUM(CASE WHEN tc.type = 'expense' THEN pt.amount ELSE 0 END), 0) AS expense
			FROM project_transactions pt
			JOIN transaction_categories tc ON tc.id = pt.category_id AND tc.deleted_at IS NULL
			JOIN projects p ON p.id = pt.project_id AND p.deleted_at IS NULL AND p.status != 'closed'
			WHERE pt.deleted_at IS NULL
		`).Scan(&r).Error
	}
	if err != nil {
		return nil, err
	}

	net := r.Income - r.Expense
	return &dashboard.PLSummary{
		Income:  fmt.Sprintf("%.2f", r.Income),
		Expense: fmt.Sprintf("%.2f", r.Expense),
		Net:     fmt.Sprintf("%.2f", net),
	}, nil
}
