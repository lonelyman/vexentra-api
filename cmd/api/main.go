package main

import (
	"log"
	"os"

	"vexentra-api/internal/bootstrap"
	"vexentra-api/internal/config"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env automatically in non-production environments.
	// In production, env vars are injected by the orchestrator (Docker/K8s).
	if os.Getenv("API_ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("[info] .env file not found, relying on system environment variables")
		}
	}
	// 1. โหลด Config พร้อมเช็ค Error ทันที (Fail-Fast)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Configuration Error: %v", err)
	}

	// 2. ประกอบร่าง App (DI & Infrastructure Setup)
	// ฟังก์ชันนี้จะจัดการทั้ง Fiber, DB, และ Redis ภายในตัวเดียว
	app, err := bootstrap.InitializeApp(cfg)
	if err != nil {
		log.Fatalf("❌ App Initialization Failed: %v", err)
	}

	// 3. รันระบบ (จัดการ Port Listening และ Graceful Shutdown)
	if err := app.Run(); err != nil {
		log.Fatalf("❌ Server crashed: %v", err)
	}
}
