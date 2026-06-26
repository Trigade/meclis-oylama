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
	"meclis-oylama/backend/internal/session"
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

	// Migration'ları otomatik çalıştır
	if err := db.RunMigrations(database, "internal/db/migrations"); err != nil {
		log.Fatalf("Migration hatası: %v", err)
	}

	// Aktif oturumu yükle
	activeSess := session.NewActiveSession()
	if err := activeSess.LoadFromDB(database); err != nil {
		log.Printf("Aktif oturum yüklenemedi: %v", err)
	}

	// Servisler
	attendanceSvc := attendance.NewService(database)
	auditSvc := auth.NewAuditService(database)
	hub := voting.NewHub()
	votingSvc := voting.NewService(database, hub)

	// Handler'lar
	membersHandler := auth.NewMembersHandler(database, auditSvc)
	auditHandler := auth.NewAuditHandler(database)
	meetingHandler := auth.NewMeetingHandler(database, auditSvc, activeSess)
	komisyonHandler := auth.NewKomisyonHandler(database, auditSvc)
	attendanceHandler := attendance.NewHandler(attendanceSvc, cfg.BridgeSecret, activeSess.Get)
	authHandler := auth.NewHandler(database, attendanceSvc, activeSess.Get, auditSvc)
	votingHandler := voting.NewHandler(votingSvc, hub, attendanceSvc, activeSess.Get)
	huzurHandler := auth.NewHuzurHandler(database, activeSess.Get)
	salonHandler := auth.NewSalonHandler(database, activeSess.Get)
	// WebSocket hub'ı arka planda başlat
	go hub.Run()

	// Router
	r := gin.Default()

	r.Static("/app", "../frontend")

	r.Use(func(c *gin.Context) {
		// Tarayıcının çerezleri engellemesini kökten çözüyoruz
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
		}

		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Cookie")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")

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
		protected.GET("/voting/active", votingHandler.GetActive)
		protected.GET("/voting/recent", votingHandler.GetRecent)
		protected.GET("/attendance/present", attendanceHandler.GetPresent)
		protected.GET("/members", membersHandler.List)
		protected.GET("/reports/votings", votingHandler.GetReport)
		protected.GET("/reports/votings/:id/detail", votingHandler.GetVoteDetail)
		protected.GET("/huzur/list", huzurHandler.List)
		protected.POST("/huzur/settings", huzurHandler.SaveSettings)
		protected.GET("/huzur/roles", huzurHandler.GetRoles)
		protected.POST("/huzur/member-role", huzurHandler.SetMemberRole)
		protected.GET("/meetings", meetingHandler.List)
		protected.POST("/meetings", meetingHandler.Create)
		protected.POST("/meetings/:id/start", meetingHandler.Start)
		protected.POST("/meetings/:id/end", meetingHandler.End)
		protected.GET("/meetings/active", meetingHandler.GetActive)
		protected.POST("/members", membersHandler.Create)
		protected.PUT("/members/:id", membersHandler.Update)
		protected.DELETE("/members/:id", membersHandler.Delete)
		protected.POST("/members/:id/reset-password", membersHandler.ResetPassword)
		protected.GET("/salon/seats", salonHandler.GetSeats)
		protected.GET("/audit", auditHandler.List)
		protected.GET("/komisyonlar", komisyonHandler.List)
		protected.POST("/komisyonlar", komisyonHandler.Create)
		protected.POST("/komisyonlar/:id/end", komisyonHandler.End)
		protected.GET("/komisyonlar/:id/uyeler", komisyonHandler.GetUyeler)
		protected.POST("/komisyonlar/:id/uyeler", komisyonHandler.AddUye)
		protected.DELETE("/komisyonlar/:id/uyeler/:mid", komisyonHandler.RemoveUye)
		protected.GET("/komisyonlar/:id/kararlar", komisyonHandler.GetKararlar)
		protected.POST("/komisyonlar/:id/kararlar", komisyonHandler.AddKarar)
		protected.DELETE("/komisyonlar/:id/kararlar/:kid", komisyonHandler.DeleteKarar)
	}

	log.Printf("Sunucu başlatılıyor → :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Sunucu başlatılamadı: %v", err)
	}
}
