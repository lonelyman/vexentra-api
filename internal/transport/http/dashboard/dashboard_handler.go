package dashboardhdl

import (
	"vexentra-api/internal/modules/dashboard/dashboardsvc"
	"vexentra-api/internal/transport/http/presenter"
	"vexentra-api/pkg/auth"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/gofiber/fiber/v3"
)

type DashboardHandler struct {
	svc    dashboardsvc.DashboardService
	logger logger.Logger
}

func NewDashboardHandler(svc dashboardsvc.DashboardService, l logger.Logger) *DashboardHandler {
	if l == nil {
		l = logger.Get()
	}
	return &DashboardHandler{svc: svc, logger: l}
}

// GetStats — GET /api/v1/dashboard/stats
func (h *DashboardHandler) GetStats(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}

	stats, svcErr := h.svc.GetStats(c.Context(), caller)
	if svcErr != nil {
		h.logger.Error("DASHBOARD_STATS_ERROR", svcErr)
		return custom_errors.New(500, "INTERNAL_ERROR", "ไม่สามารถดึงข้อมูล Dashboard ได้")
	}

	return presenter.RenderItem(c, NewStatsResponse(stats))
}
