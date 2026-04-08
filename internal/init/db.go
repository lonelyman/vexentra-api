package init

import (
	"context"
	"fmt"
	"log/slog" // 🆕 เพิ่ม
	"time"
	"vexentra-api/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type gormLoggerAdapter struct {
	l *slog.Logger // 🆕 เก็บ logger ไว้
}

func (l *gormLoggerAdapter) LogMode(level logger.LogLevel) logger.Interface { return l }
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
	l.l.Debug("SQL_TRACE",
		slog.Duration("elapsed", elapsed),
		slog.Int64("rows", rows),
		slog.String("sql", sql),
		slog.Any("error", err),
	)
}

// NewDatabaseConnection รับ logger เพิ่มเข้ามา
func NewDatabaseConnection(cfg config.PostgresConfig, l *slog.Logger) (*gorm.DB, error) {
	dsn := cfg.BuildDSN()

	gormConfig := &gorm.Config{
		Logger: &gormLoggerAdapter{l: l}, // 🆕 ส่ง logger เข้าไป
	}
	// ... (ส่วนที่เหลือคงเดิม)
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	return db, nil
}
