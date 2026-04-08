package logger

import (
	"log/slog"
	"os"
)

func New() *slog.Logger {
	// ใน Production ควรใช้ slog.NewJSONHandler
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler)
}
