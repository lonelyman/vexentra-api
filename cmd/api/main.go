package main

import (
	"log"
	"vexentra-api/internal/config"
	appinit "vexentra-api/internal/init"
)

func main() {
	// 1. โหลด Config พร้อมเช็ค Error ทันที (Fail-Fast)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Configuration Error: %v", err)
	}

	// 2. ประกอบร่าง App (DI & Infrastructure Setup)
	// ฟังก์ชันนี้จะจัดการทั้ง Fiber, DB, และ Redis ภายในตัวเดียว
	app, err := appinit.InitializeApp(cfg)
	if err != nil {
		log.Fatalf("❌ App Initialization Failed: %v", err)
	}

	// 3. รันระบบ (จัดการ Port Listening และ Graceful Shutdown)
	if err := app.Run(); err != nil {
		log.Fatalf("❌ Server crashed: %v", err)
	}
}
