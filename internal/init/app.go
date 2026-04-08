package init

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
	userhdl "vexentra-api/internal/transport/http/user"
	"vexentra-api/pkg/logger" // ✅ มั่นใจว่า import pkg/logger

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type App struct {
	cfg    *config.Config
	server *fiber.App
	db     *gorm.DB
	redis  *redis.Client
	logger logger.Logger // ✅ เปลี่ยนเป็น Interface
}

func InitializeApp(cfg *config.Config) (*App, error) {
	// 0. สร้าง Logger (Interface)
	logger.Init(cfg.App.Env)
	l := logger.Get()

	// 1. ต่อ DB (ส่ง Interface เข้าไป)
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
	userRepo := pguser.NewUserRepository(db, l)
	userSvc := usersvc.NewUserService(userRepo, l)
	userHdl := userhdl.NewUserHandler(userSvc, l)

	if err := pguser.AutoMigrate(db); err != nil {
		return nil, err
	}

	// 4. สร้าง Server (ส่ง Interface เข้าไป)
	server := NewFiberServer(cfg, l)

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

	return a.server.ShutdownWithContext(ctx)
}
