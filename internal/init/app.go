package init

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vexentra-api/internal/adapters/database/postgres/pguser"
	"vexentra-api/internal/config"
	"vexentra-api/internal/modules/user/usersvc"
	userhdl "vexentra-api/internal/transport/http/user"
	"vexentra-api/pkg/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type App struct {
	cfg    *config.Config
	server *fiber.App
	db     *gorm.DB
	redis  *redis.Client
	logger *slog.Logger
}

func InitializeApp(cfg *config.Config) (*App, error) {
	l := logger.New()

	// 1. ต่อ Database (ส่ง l เข้าไปด้วย และใช้ cfg.Postgres.Primary ตามจริง)
	db, err := NewDatabaseConnection(cfg.Postgres.Primary, l)
	if err != nil {
		return nil, err
	}

	// 2. ต่อ Redis
	rdb, err := NewRedisConnection(cfg.Redis)
	if err != nil {
		return nil, err
	}

	// 3. DI
	userRepo := pguser.NewUserRepository(db)
	userSvc := usersvc.NewUserService(userRepo)
	userHdl := userhdl.NewUserHandler(userSvc)

	// 4. Fiber Server (ส่ง l เข้าไป และใช้ cfg.App.AppPort ในอนาคต)
	server := NewFiberServer(cfg, l)

	// Route Registration
	api := server.Group("/api/v1")
	api.Post("/users/register", userHdl.Register)

	return &App{
		cfg:    cfg,
		server: server,
		db:     db,
		redis:  rdb,
		logger: l,
	}, nil
}

func (a *App) Run() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// ✅ ใช้ a.cfg.App.AppPort ตามความจริงใน config.go
		addr := fmt.Sprintf(":%s", a.cfg.App.AppPort)
		a.logger.Info("🚀 Vexentra API Starting", "address", addr)
		if err := a.server.Listen(addr); err != nil {
			a.logger.Error("❌ Server Listen Error", "error", err)
		}
	}()

	<-stop
	a.logger.Info("🧹 Cleaning up resources...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return a.server.ShutdownWithContext(ctx)
}
