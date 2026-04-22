package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vexentra-api/internal/adapters/database/postgres/pgperson"
	"vexentra-api/internal/adapters/database/postgres/pgproject"
	"vexentra-api/internal/adapters/database/postgres/pgsocialplatform"
	"vexentra-api/internal/adapters/database/postgres/pgtask"
	"vexentra-api/internal/adapters/database/postgres/pgtxcategory"
	"vexentra-api/internal/adapters/database/postgres/pguser"
	"vexentra-api/internal/config"
	"vexentra-api/internal/modules/dashboard/dashboardsvc"
	"vexentra-api/internal/modules/project/projectsvc"
	"vexentra-api/internal/modules/socialplatform/platformsvc"
	"vexentra-api/internal/modules/task/tasksvc"
	"vexentra-api/internal/modules/txcategory/txcategorysvc"
	"vexentra-api/internal/modules/user/usersvc"
	"vexentra-api/internal/transport/http"
	authhdl "vexentra-api/internal/transport/http/auth"
	dashboardhdl "vexentra-api/internal/transport/http/dashboard"
	healthhdl "vexentra-api/internal/transport/http/health"
	projecthdl "vexentra-api/internal/transport/http/project"
	socialplatformhdl "vexentra-api/internal/transport/http/socialplatform"
	taskhdl "vexentra-api/internal/transport/http/task"
	txcategoryhdl "vexentra-api/internal/transport/http/txcategory"
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
	projectRepo := pgproject.NewProjectRepository(db, l)
	memberRepo := pgproject.NewProjectMemberRepository(db, l)
	txRepo := pgproject.NewProjectTransactionRepository(db, l)
	categoryRepo := pgtxcategory.NewTransactionCategoryRepository(db, l)

	userSvc := usersvc.NewUserService(db, userRepo, personRepo, authSvc, l)
	profileSvc := usersvc.NewProfileService(userRepo, profileRepo, socialPlatformRepo, l)
	socialPlatformSvc := platformsvc.NewSocialPlatformService(socialPlatformRepo, l)
	projectSvc := projectsvc.NewProjectService(db, projectRepo, memberRepo, cfg.App.ProjectCodePrefix, l)
	memberSvc := projectsvc.NewMemberService(projectSvc, memberRepo, l)
	txSvc := projectsvc.NewTransactionService(projectSvc, memberRepo, txRepo, categoryRepo, l)
	categorySvc := txcategorysvc.NewTransactionCategoryService(categoryRepo, l)

	taskRepo := pgtask.NewTaskRepository(db, l)
	taskSvc := tasksvc.New(projectSvc, memberRepo, taskRepo, l)
	dashboardSvc := dashboardsvc.New(db, l)

	userHdl := userhdl.NewUserHandler(userSvc, l)
	profileHdl := userhdl.NewProfileHandler(profileSvc, cfg.App.ShowcasePersonID, l)
	socialPlatformHdl := socialplatformhdl.NewSocialPlatformHandler(socialPlatformSvc, l)
	authHdl := authhdl.NewAuthHandler(userSvc, authSvc, cfg.App.Env, l)
	healthHdl := healthhdl.NewHealthHandler(db, rdb)
	projectHdl := projecthdl.NewProjectHandler(projectSvc, l)
	memberHdl := projecthdl.NewMemberHandler(memberSvc, l)
	txHdl := projecthdl.NewTransactionHandler(txSvc, l)
	txCategoryHdl := txcategoryhdl.NewCategoryHandler(categorySvc, l)
	dashboardHdl := dashboardhdl.NewDashboardHandler(dashboardSvc, l)
	taskHdl := taskhdl.NewTaskHandler(taskSvc, l)

	http.SetupRouter(server, http.Handlers{
		User:           userHdl,
		Profile:        profileHdl,
		SocialPlatform: socialPlatformHdl,
		Auth:           authHdl,
		Health:         healthHdl,
		Project:        projectHdl,
		Member:         memberHdl,
		Transaction:    txHdl,
		TxCategory:     txCategoryHdl,
		Dashboard:      dashboardHdl,
		Task:           taskHdl,
		AuthSvc:        authSvc,
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
