package auth

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MeetingHandler struct {
	db         *sql.DB
	audit      *AuditService
	activeSess interface{ Set(string) }
}

func NewMeetingHandler(db *sql.DB, audit *AuditService, activeSess interface{ Set(string) }) *MeetingHandler {
	return &MeetingHandler{db: db, audit: audit, activeSess: activeSess}
}

type Meeting struct {
	ID        int     `json:"id"`
	Title     string  `json:"title"`
	MeetingNo string  `json:"meeting_no"`
	Status    string  `json:"status"`
	PlannedAt *string `json:"planned_at"`
	StartedAt *string `json:"started_at"`
	EndedAt   *string `json:"ended_at"`
}

// GET /api/meetings
func (h *MeetingHandler) List(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, title, COALESCE(meeting_no,''), status,
		       planned_at::text, started_at::text, ended_at::text
		FROM meetings ORDER BY id DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "oturumlar alınamadı"})
		return
	}
	defer rows.Close()

	var list []Meeting
	for rows.Next() {
		var m Meeting
		rows.Scan(&m.ID, &m.Title, &m.MeetingNo, &m.Status,
			&m.PlannedAt, &m.StartedAt, &m.EndedAt)
		list = append(list, m)
	}
	if list == nil {
		list = []Meeting{}
	}
	c.JSON(http.StatusOK, list)
}

// POST /api/meetings
func (h *MeetingHandler) Create(c *gin.Context) {
	var req struct {
		Title     string `json:"title" binding:"required"`
		MeetingNo string `json:"meeting_no"`
		PlannedAt string `json:"planned_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var m Meeting
	err := h.db.QueryRow(`
		INSERT INTO meetings (title, meeting_no, status, planned_at)
		VALUES ($1, $2, 'planned', $3::timestamptz)
		RETURNING id, title, COALESCE(meeting_no,''), status, planned_at::text
	`, req.Title, req.MeetingNo, nullStr(req.PlannedAt)).Scan(
		&m.ID, &m.Title, &m.MeetingNo, &m.Status, &m.PlannedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "oturum oluşturulamadı"})
		return
	}
	c.JSON(http.StatusOK, m)
}

// POST /api/meetings/:id/start
func (h *MeetingHandler) Start(c *gin.Context) {
	id := c.Param("id")

	// Başka aktif oturum var mı?
	var count int
	h.db.QueryRow(`SELECT COUNT(*) FROM meetings WHERE status = 'active'`).Scan(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "zaten aktif bir oturum var"})
		return
	}

	var m Meeting
	err := h.db.QueryRow(`
		UPDATE meetings SET status = 'active', started_at = NOW()
		WHERE id = $1 AND status = 'planned'
		RETURNING id, title, COALESCE(meeting_no,''), status, started_at::text
	`, id).Scan(&m.ID, &m.Title, &m.MeetingNo, &m.Status, &m.StartedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "oturum başlatılamadı"})
		return
	}

	h.activeSess.Set(fmt.Sprintf("meeting-%d", m.ID))

	h.audit.Log(AuditEntry{
		Action:     "meeting_start",
		TargetType: "meeting",
		TargetID:   id,
		Detail:     m.Title,
		IP:         c.ClientIP(),
	})
	c.JSON(http.StatusOK, m)
}

// POST /api/meetings/:id/end
func (h *MeetingHandler) End(c *gin.Context) {
	id := c.Param("id")

	var m Meeting
	err := h.db.QueryRow(`
		UPDATE meetings SET status = 'ended', ended_at = NOW()
		WHERE id = $1 AND status = 'active'
		RETURNING id, title, COALESCE(meeting_no,''), status, ended_at::text
	`, id).Scan(&m.ID, &m.Title, &m.MeetingNo, &m.Status, &m.EndedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "oturum kapatılamadı"})
		return
	}

	h.activeSess.Set("meeting-default")

	h.audit.Log(AuditEntry{
		Action:     "meeting_end",
		TargetType: "meeting",
		TargetID:   id,
		Detail:     m.Title,
		IP:         c.ClientIP(),
	})
	c.JSON(http.StatusOK, m)
}

// GET /api/meetings/active
func (h *MeetingHandler) GetActive(c *gin.Context) {
	var m Meeting
	err := h.db.QueryRow(`
		SELECT id, title, COALESCE(meeting_no,''), status, started_at::text
		FROM meetings WHERE status = 'active' LIMIT 1
	`).Scan(&m.ID, &m.Title, &m.MeetingNo, &m.Status, &m.StartedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, gin.H{"active": false})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "aktif oturum alınamadı"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"active": true, "meeting": m})
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
