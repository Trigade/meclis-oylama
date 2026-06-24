package attendance

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service   *Service
	secret    string
	meetingID string
}

func NewHandler(service *Service, secret, meetingID string) *Handler {
	return &Handler{
		service:   service,
		secret:    secret,
		meetingID: meetingID,
	}
}

type eventRequest struct {
	MemberID  int     `json:"member_id" binding:"required"`
	EventType string  `json:"event_type" binding:"required,oneof=entry exit"`
	Timestamp float64 `json:"timestamp"`
}

// POST /api/bridge/attendance
func (h *Handler) HandleEvent(c *gin.Context) {
	// Köprü servisi gizli anahtarını doğrula
	if c.GetHeader("X-Bridge-Secret") != h.secret {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "geçersiz bridge secret"})
		return
	}

	var req eventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var err error
	switch req.EventType {
	case "entry":
		err = h.service.RecordEntry(h.meetingID, req.MemberID)
	case "exit":
		err = h.service.RecordExit(h.meetingID, req.MemberID)
	}

	if err != nil {
		log.Printf("Yoklama hatası: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "kayıt başarısız"})
		return
	}

	count, _ := h.service.CountPresent(h.meetingID)
	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"present_count": count,
	})
}

// GET /api/attendance/count
func (h *Handler) GetCount(c *gin.Context) {
	count, err := h.service.CountPresent(h.meetingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sayım alınamadı"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"present_count": count})
}
