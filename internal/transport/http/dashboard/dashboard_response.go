package dashboardhdl

import (
	"vexentra-api/internal/modules/dashboard"
)

type StatusCountResponse struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type DeadlineProjectResponse struct {
	ID          string `json:"id"`
	ProjectCode string `json:"project_code"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	DeadlineAt  string `json:"deadline_at"`
}

type PLSummaryResponse struct {
	Income  string `json:"income"`
	Expense string `json:"expense"`
	Net     string `json:"net"`
}

type StatsResponse struct {
	StatusCounts      []StatusCountResponse     `json:"status_counts"`
	UpcomingDeadlines []DeadlineProjectResponse `json:"upcoming_deadlines"`
	PL                PLSummaryResponse         `json:"pl"`
}

func NewStatsResponse(s *dashboard.Stats) StatsResponse {
	counts := make([]StatusCountResponse, len(s.StatusCounts))
	for i, c := range s.StatusCounts {
		counts[i] = StatusCountResponse{
			Status: string(c.Status),
			Count:  c.Count,
		}
	}

	deadlines := make([]DeadlineProjectResponse, len(s.UpcomingDeadlines))
	for i, d := range s.UpcomingDeadlines {
		deadlines[i] = DeadlineProjectResponse{
			ID:          d.ID.String(),
			ProjectCode: d.ProjectCode,
			Name:        d.Name,
			Status:      string(d.Status),
			DeadlineAt:  d.DeadlineAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return StatsResponse{
		StatusCounts:      counts,
		UpcomingDeadlines: deadlines,
		PL: PLSummaryResponse{
			Income:  s.PL.Income,
			Expense: s.PL.Expense,
			Net:     s.PL.Net,
		},
	}
}
