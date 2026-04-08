// vexentra-api/internal/bootstrap/redis.go
package bootstrap

import (
	"context"
	"fmt"
	"time"
	"vexentra-api/internal/config"

	"github.com/redis/go-redis/v9"
)

// NewRedisConnection สร้างการเชื่อมต่อและตรวจสอบสถานะด้วยการ Ping
func NewRedisConnection(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		// ปรับจูน Pool สำหรับงานหนัก
		PoolSize:     20,
		MinIdleConns: 5,
	})

	// ตรวจสอบการเชื่อมต่อจริง (Fail-Fast)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return rdb, nil
}
