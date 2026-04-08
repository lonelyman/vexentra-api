package healthhdl

import (
	"context"
	"time"

	"vexentra-api/internal/transport/http/presenter"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type healthStatus string

const (
	statusHealthy   healthStatus = "healthy"
	statusUnhealthy healthStatus = "unhealthy"
)

type HealthHandler struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewHealthHandler(db *gorm.DB, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, rdb: rdb}
}

// Live handles GET /health/live
// Only checks that the process is running — never touches DB or Redis.
// Used by Kubernetes livenessProbe to decide whether to restart the pod.
func (h *HealthHandler) Live(c fiber.Ctx) error {
	return presenter.RenderItem(c, fiber.Map{"status": "alive"})
}

// Ready handles GET /health/ready
// Pings all downstream dependencies and reports per-component status.
// Used by Kubernetes readinessProbe to decide whether to route traffic to the pod.
func (h *HealthHandler) Ready(c fiber.Ctx) error {
	type ReadinessResponse struct {
		Status     healthStatus            `json:"status"`
		Components map[string]healthStatus `json:"components"`
	}

	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	components := map[string]healthStatus{}
	overall := statusHealthy

	// Check Database
	if sqlDB, err := h.db.DB(); err == nil {
		if err := sqlDB.PingContext(ctx); err == nil {
			components["database"] = statusHealthy
		} else {
			components["database"] = statusUnhealthy
			overall = statusUnhealthy
		}
	} else {
		components["database"] = statusUnhealthy
		overall = statusUnhealthy
	}

	// Check Redis
	if err := h.rdb.Ping(ctx).Err(); err == nil {
		components["redis"] = statusHealthy
	} else {
		components["redis"] = statusUnhealthy
		overall = statusUnhealthy
	}

	statusCode := fiber.StatusOK
	if overall == statusUnhealthy {
		statusCode = fiber.StatusServiceUnavailable
	}

	return presenter.RenderItem(c, ReadinessResponse{
		Status:     overall,
		Components: components,
	}, statusCode)
}
