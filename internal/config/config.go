// vexentra-api/internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config รวบรวมทุกส่วนของระบบ (เน้นสัดส่วนตามแนวทางของนายท่าน)
type Config struct {
	App      AppConfig
	Postgres PostgresDbs
	Redis    RedisConfig
}

type AppConfig struct {
	Env      string
	AppPort  string
	Timezone string
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
			Env:      mustGetEnv("API_ENV", &missingKeys),
			AppPort:  mustGetEnv("API_PORT", &missingKeys),
			Timezone: getEnv("API_TIMEZONE", "Asia/Bangkok"), // Timezone ให้มี default ได้
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
