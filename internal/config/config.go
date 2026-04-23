// vexentra-api/internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config รวบรวมทุกส่วนของระบบ (เน้นสัดส่วนตามแนวทางของนายท่าน)
type Config struct {
	App      AppConfig
	JWT      JWTConfig
	Mailer   MailerConfig
	Postgres PostgresDbs
	Redis    RedisConfig
}

type AppConfig struct {
	Env                string
	AppPort            string
	Timezone           string
	WebBaseURL         string   // base URL for frontend links used in email templates
	CORSAllowedOrigins []string // comma-separated via API_CORS_ALLOWED_ORIGINS
	ShowcasePersonID   string   // optional: fixed person ID for public showcase endpoint
	ProjectCodePrefix  string   // uppercase alphabetic prefix for PREFIX-YYYY-NNNN project codes
}

type PostgresDbs struct {
	Primary PostgresConfig
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// BuildDSN สร้าง URI สำหรับ GORM (รูปแบบเดียวกับโปรเจกต์เดิมของนายท่าน)
func (p PostgresConfig) BuildDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.DBName, p.SSLMode)
}

type JWTConfig struct {
	AccessSecret  string
	AccessExpiry  time.Duration // e.g. "15m"
	RefreshSecret string
	RefreshExpiry time.Duration // e.g. "168h" (7 days)
	Issuer        string
}

type MailerConfig struct {
	Host     string
	Port     int
	Name     string
	Username string
	Password string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// LoadConfig ทำหน้าที่โหลดค่าและตรวจสอบความครบถ้วน (Fail-Fast)
func LoadConfig() (*Config, error) {
	var missingKeys []string

	cfg := &Config{
		App: AppConfig{
			Env:                mustGetEnv("API_ENV", &missingKeys),
			AppPort:            mustGetEnv("API_PORT", &missingKeys),
			Timezone:           getEnv("API_TIMEZONE", "Asia/Bangkok"),
			WebBaseURL:         getEnv("APP_WEB_URL", "http://localhost:3005"),
			CORSAllowedOrigins: getEnvAsSlice("API_CORS_ALLOWED_ORIGINS", nil),
			ShowcasePersonID:   getEnv("APP_SHOWCASE_PERSON_ID", ""),
			ProjectCodePrefix:  strings.ToUpper(getEnv("APP_PROJECT_CODE_PREFIX", "VX")),
		},
		Postgres: PostgresDbs{
			Primary: PostgresConfig{
				Host:     mustGetEnv("POSTGRES_PRIMARY_HOST", &missingKeys),
				Port:     mustGetEnv("POSTGRES_PRIMARY_PORT", &missingKeys),
				User:     mustGetEnv("POSTGRES_PRIMARY_USER", &missingKeys),
				Password: mustGetEnv("POSTGRES_PRIMARY_PASSWORD", &missingKeys),
				DBName:   mustGetEnv("POSTGRES_PRIMARY_NAME", &missingKeys),
				SSLMode:  mustGetEnv("POSTGRES_PRIMARY_SSL_MODE", &missingKeys),
			},
		},
		JWT: JWTConfig{
			AccessSecret:  mustGetEnv("JWT_ACCESS_SECRET", &missingKeys),
			AccessExpiry:  mustGetEnvAsDuration("JWT_ACCESS_EXPIRY", &missingKeys),
			RefreshSecret: mustGetEnv("JWT_REFRESH_SECRET", &missingKeys),
			RefreshExpiry: mustGetEnvAsDuration("JWT_REFRESH_EXPIRY", &missingKeys),
			Issuer:        getEnv("JWT_ISSUER", "vexentra-api"),
		},
		Mailer: MailerConfig{
			Host:     getEnv("MAILER_HOST", ""),
			Port:     getEnvAsInt("MAILER_PORT", 587, &missingKeys),
			Name:     getEnv("MAILER_NAME", "Vexentra"),
			Username: getEnv("MAILER_USERNAME", ""),
			Password: getEnv("MAILER_PASSWORD", ""),
		},
		Redis: RedisConfig{
			Host:     mustGetEnv("REDIS_HOST", &missingKeys),
			Port:     mustGetEnv("REDIS_PORT", &missingKeys),
			Password: getEnv("REDIS_PASSWORD", ""), // Password อนุญาตให้ว่างได้
			DB:       getEnvAsInt("REDIS_DB", 0, &missingKeys),
		},
	}

	// ถ้ามี Key สำคัญหายไป ให้รวบรวมมาบอกทีเดียว
	if len(missingKeys) > 0 {
		return nil, fmt.Errorf("❌ configuration failed! missing keys: [%s]", strings.Join(missingKeys, ", "))
	}

	return cfg, nil
}

// --- Internal Helpers ---

func mustGetEnv(key string, missing *[]string) string {
	val := os.Getenv(key)
	if val == "" {
		*missing = append(*missing, key)
	}
	return val
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int, missing *[]string) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		*missing = append(*missing, fmt.Sprintf("%s (must be integer)", key))
		return 0
	}
	return val
}

func mustGetEnvAsDuration(key string, missing *[]string) time.Duration {
	valStr := os.Getenv(key)
	if valStr == "" {
		*missing = append(*missing, key)
		return 0
	}
	d, err := time.ParseDuration(valStr)
	if err != nil {
		*missing = append(*missing, fmt.Sprintf("%s (must be duration e.g. 15m, 168h)", key))
		return 0
	}
	return d
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	parts := strings.Split(val, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
