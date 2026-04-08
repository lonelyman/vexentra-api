package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vexentra-api/internal/adapters/database/postgres/pguser"
	"vexentra-api/internal/config"
	"vexentra-api/internal/modules/user/usersvc"
	"vexentra-api/internal/transport/http"
	authhdl "vexentra-api/internal/transport/http/auth"
	healthhdl "vexentra-api/internal/transport/http/health"
	userhdl "vexentra-api/internal/transport/http/user"
	"vexentra-api/pkg/auth"
	"vexentra-api/pkg/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type App struct {
	cfg     *config.Config
	server  *fiber.App
	db      *gorm.DB
	redis   *redis.Client
	logger  logger.Logger
	authSvc auth.AuthService
}

func InitializeApp(cfg *config.Config) (*App, error) {
	logger.Init(cfg.App.Env)
	l := logger.Get()

	db, err := NewDatabaseConnection(cfg.Postgres.Primary, l)
	if err != nil {
		return nil, err
	}

	rdb, err := NewRedisConnection(cfg.Redis)
	if err != nil {
		return nil, err
	}

	// Run Migrations (ควรจัดกลุ่มรวมกัน)
	if err := pguser.AutoMigrate(db); err != nil {
		return nil, err
	}

	server := NewFiberServer(cfg, l)

	// DI Section
	authSvc := auth.NewAuthService(cfg.JWT)
	userRepo := pguser.NewUserRepository(db, l)
	userSvc := usersvc.NewUserService(userRepo, authSvc, l)
	userHdl := userhdl.NewUserHandler(userSvc, l)
	authHdl := authhdl.NewAuthHandler(userSvc, authSvc, l)
	healthHdl := healthhdl.NewHealthHandler(db, rdb)

	http.SetupRouter(server, http.Handlers{
		User:    userHdl,
		Auth:    authHdl,
		Health:  healthHdl,
		AuthSvc: authSvc,
		// Order: orderHdl,
	})

	return &App{
		cfg:     cfg,
		server:  server,
		db:      db,
		redis:   rdb,
		logger:  l,
		authSvc: authSvc,
	}, nil
}

func (a *App) Run() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%s", a.cfg.App.AppPort)
		a.logger.Info("🚀 Vexentra API Starting", "address", addr)
		if err := a.server.Listen(addr); err != nil {
			a.logger.Error("❌ Server Listen Error", err)
		}
	}()

	<-stop
	a.logger.Info("🧹 Cleaning up resources...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ✅ Graceful Shutdown ครบวงจร
	if sqlDB, err := a.db.DB(); err == nil {
		sqlDB.Close()
		a.logger.Info("📦 Database connection closed.")
	}
	a.redis.Close()
	a.logger.Info("🔑 Redis connection closed.")

	return a.server.ShutdownWithContext(ctx)
}
