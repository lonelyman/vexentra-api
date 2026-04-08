package init

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vexentra-api/internal/config"
	"vexentra-api/pkg/logger" // ✅ Import Interface ของเรา

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLog "gorm.io/gorm/logger" // Alias เพื่อไม่ให้ชนกับ pkg/logger
)

type gormLoggerAdapter struct {
	l *slog.Logger
}

func (l *gormLoggerAdapter) LogMode(level gormLog.LogLevel) gormLog.Interface { return l }
func (l *gormLoggerAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	l.l.Info(fmt.Sprintf(msg, data...))
}
func (l *gormLoggerAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.l.Warn(fmt.Sprintf(msg, data...))
}
func (l *gormLoggerAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	l.l.Error(fmt.Sprintf(msg, data...))
}
func (l *gormLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rows := fc()
	elapsed := time.Since(begin)
	if err != nil {
		l.l.Error("SQL_TRACE_ERROR", slog.Duration("elapsed", elapsed), slog.Int64("rows", rows), slog.String("sql", sql), slog.Any("error", err))
	} else {
		l.l.Debug("SQL_TRACE", slog.Duration("elapsed", elapsed), slog.Int64("rows", rows), slog.String("sql", sql))
	}
}

// ✅ แก้ไข: รับ logger.Logger (Interface)
func NewDatabaseConnection(cfg config.PostgresConfig, l logger.Logger) (*gorm.DB, error) {
	if l == nil {
		l = logger.Get()
	}
	dsn := cfg.BuildDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: &gormLoggerAdapter{l: l.GetSlog()},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
