package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
)

var (
	L    Logger
	once sync.Once
)

// LevelSuccess กำหนดระดับความสำเร็จ (มากกว่า Info แต่น้อยกว่า Warn)
const LevelSuccess = slog.Level(1)

// Logger คือ Interface สัญญาใจที่ทุกที่เรียกใช้
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Success(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, err error, args ...any)
	Dump(msg string, data any)
	GetSlog() *slog.Logger
}

type appLogger struct {
	l *slog.Logger
}

func (a *appLogger) Debug(msg string, args ...any) { a.l.Debug(msg, args...) }
func (a *appLogger) Info(msg string, args ...any)  { a.l.Info(msg, args...) }
func (a *appLogger) Success(msg string, args ...any) {
	a.l.Log(context.Background(), LevelSuccess, msg, args...)
}
func (a *appLogger) Warn(msg string, args ...any) { a.l.Warn(msg, args...) }
func (a *appLogger) Error(msg string, err error, args ...any) {
	args = append(args, "error", err)
	a.l.Error(msg, args...)
}
func (a *appLogger) Dump(msg string, data any) {
	// Dump จะทำงานเฉพาะตอนตั้ง Level เป็น Debug เท่านั้น
	a.l.Debug(msg, "data_dump", data)
}
func (a *appLogger) GetSlog() *slog.Logger { return a.l }

// Init ทำหน้าที่ตั้งค่า Logger ครั้งเดียวตอนเริ่มแอป (เรียกจาก InitializeApp)
func Init(env string) {
	once.Do(func() {
		// ✅ ตั้งเป็น LevelInfo เพื่อกรอง SQL Trace (Debug) ทิ้งไปตามที่นายท่านต้องการ
		opts := slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
		}

		var handler slog.Handler
		if env == "production" {
			handler = slog.NewJSONHandler(os.Stdout, &opts)
		} else {
			// ✅ ใช้ PrettyHandler ตัวสวยงามตอนพัฒนา
			handler = NewPrettyHandler(os.Stdout, opts)
		}
		L = &appLogger{l: slog.New(handler)}
	})
}

// Get สำหรับดึง Global Logger ไปใช้งาน
func Get() Logger {
	if L == nil {
		Init("development")
	}
	return L
}

// --- PrettyHandler Section (ยุบรวมมาไว้ที่นี่แล้วค่ะ) ---

type PrettyHandler struct {
	opts slog.HandlerOptions
	out  io.Writer
}

func NewPrettyHandler(out io.Writer, opts slog.HandlerOptions) *PrettyHandler {
	return &PrettyHandler{out: out, opts: opts}
}

func (h *PrettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String()
	color := "\033[97m" // White
	emoji := "ℹ️ "

	switch r.Level {
	case slog.LevelDebug:
		color = "\033[94m"
		emoji = "🐛 "
	case LevelSuccess:
		level = "SUCCESS"
		color = "\033[92m"
		emoji = "✅ "
	case slog.LevelWarn:
		color = "\033[93m"
		emoji = "⚠️ "
	case slog.LevelError:
		color = "\033[91m"
		emoji = "❌ "
	}

	timeStr := r.Time.Format("15:04:05")
	fmt.Fprintf(h.out, "%s %s%s %-7s\033[0m %s", timeStr, color, emoji, level, r.Message)

	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "data_dump" {
			b, _ := json.MarshalIndent(a.Value.Any(), "", "  ")
			fmt.Fprintf(h.out, "\n\033[95m🔍 DUMP:\033[0m\n%s", string(b))
		} else {
			fmt.Fprintf(h.out, " \033[90m%s=\033[0m%v", a.Key, a.Value)
		}
		return true
	})

	fmt.Fprintf(h.out, "\n")
	return nil
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *PrettyHandler) WithGroup(name string) slog.Handler       { return h }
