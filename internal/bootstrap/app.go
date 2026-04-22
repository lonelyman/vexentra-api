package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vexentra-api/internal/adapters/database/postgres/pgperson"
	"vexentra-api/internal/adapters/database/postgres/pgsocialplatform"
	"vexentra-api/internal/adapters/database/postgres/pguser"
	"vexentra-api/internal/config"
	"vexentra-api/internal/modules/socialplatform/platformsvc"
	"vexentra-api/internal/modules/user/usersvc"
	"vexentra-api/internal/transport/http"
	authhdl "vexentra-api/internal/transport/http/auth"
	healthhdl "vexentra-api/internal/transport/http/health"
	socialplatformhdl "vexentra-api/internal/transport/http/socialplatform"
	userhdl "vexentra-api/internal/transport/http/user"
	"vexentra-api/pkg/auth"
	"vexentra-api/pkg/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type App struct {
	cfg     *config.Config
	db      *gorm.DB
	redis   *redis.Client
	authSvc auth.AuthService
	server  *fiber.App
	logger  logger.Logger
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

	// Schema is managed by SQL migrations under database/migrations (goose).
	// GORM is used for query/ORM only — it does not alter table structure.

	server := NewFiberServer(cfg, l)

	// DI Section
	authSvc := auth.NewAuthService(cfg.JWT)
	userRepo := pguser.NewUserRepository(db, l)
	personRepo := pgperson.NewPersonRepository(db, l)
	profileRepo := pguser.NewProfileRepository(db, l)
	socialPlatformRepo := pgsocialplatform.NewSocialPlatformRepository(db, l)
	userSvc := usersvc.NewUserService(userRepo, personRepo, authSvc, l)
	profileSvc := usersvc.NewProfileService(userRepo, profileRepo, socialPlatformRepo, l)
	socialPlatformSvc := platformsvc.NewSocialPlatformService(socialPlatformRepo, l)
	userHdl := userhdl.NewUserHandler(userSvc, l)
	profileHdl := userhdl.NewProfileHandler(profileSvc, cfg.App.ShowcasePersonID, l)
	socialPlatformHdl := socialplatformhdl.NewSocialPlatformHandler(socialPlatformSvc, l)
	authHdl := authhdl.NewAuthHandler(userSvc, authSvc, cfg.App.Env, l)
	healthHdl := healthhdl.NewHealthHandler(db, rdb)

	http.SetupRouter(server, http.Handlers{
		User:           userHdl,
		Profile:        profileHdl,
		SocialPlatform: socialPlatformHdl,
		Auth:           authHdl,
		Health:         healthHdl,
		AuthSvc:        authSvc,
		// Order: orderHdl,
	})

	return &App{
		cfg:     cfg,
		db:      db,
		redis:   rdb,
		authSvc: authSvc,
		server:  server,
		logger:  l,
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
