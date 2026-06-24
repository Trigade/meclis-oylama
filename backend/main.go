package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"meclis-oylama/backend/config"
	"meclis-oylama/backend/internal/attendance"
	"meclis-oylama/backend/internal/auth"
	"meclis-oylama/backend/internal/db"
	"meclis-oylama/backend/internal/voting"

	"github.com/gin-gonic/gin"
)

const defaultMeetingID = "meeting-2026-001"

func main() {
	if data, err := os.ReadFile(".env"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}
	}

	cfg := config.Load()

	// Veritabanı bağlantısı
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Veritabanına bağlanılamadı: %v", err)
	}
	defer database.Close()

	// Servisler
	attendanceSvc := attendance.NewService(database)
	hub := voting.NewHub()
	votingSvc := voting.NewService(database, hub)

	// Handler'lar
	attendanceHandler := attendance.NewHandler(attendanceSvc, cfg.BridgeSecret, defaultMeetingID)
	authHandler := auth.NewHandler(database, attendanceSvc, defaultMeetingID)
	votingHandler := voting.NewHandler(votingSvc, hub, attendanceSvc, defaultMeetingID)

	// WebSocket hub'ı arka planda başlat
	go hub.Run()

	// Router
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Public rotalar
	r.POST("/api/auth/login", authHandler.Login)
	r.POST("/api/auth/logout", authHandler.Logout)

	// Köprü servisi rotası (secret ile korunur)
	r.POST("/api/bridge/attendance", attendanceHandler.HandleEvent)

	// Oturum gerektiren rotalar
	protected := r.Group("/api")
	protected.Use(auth.RequireSession())
	{
		protected.GET("/auth/me", authHandler.Me)
		protected.GET("/attendance/count", attendanceHandler.GetCount)
		protected.POST("/voting/start", votingHandler.Start)
		protected.POST("/voting/:id/vote", votingHandler.CastVote)
		protected.POST("/voting/:id/finalize", votingHandler.Finalize)
		protected.GET("/voting/ws", votingHandler.WS)
	}

	log.Printf("Sunucu başlatılıyor → :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Sunucu başlatılamadı: %v", err)
	}
}
