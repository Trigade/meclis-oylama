package attendance

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service      *Service
	secret       string
	getMeetingID func() string
}

func NewHandler(service *Service, secret string, getMeetingID func() string) *Handler {
	return &Handler{
		service:      service,
		secret:       secret,
		getMeetingID: getMeetingID,
	}
}

type eventRequest struct {
	MemberID  int     `json:"member_id" binding:"required"`
	EventType string  `json:"event_type" binding:"required,oneof=entry exit"`
	Timestamp float64 `json:"timestamp"`
}

// POST /api/bridge/attendance
func (h *Handler) HandleEvent(c *gin.Context) {
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
		err = h.service.RecordEntry(h.getMeetingID(), req.MemberID)
	case "exit":
		err = h.service.RecordExit(h.getMeetingID(), req.MemberID)
	}

	if err != nil {
		log.Printf("Yoklama hatası: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "kayıt başarısız"})
		return
	}

	count, _ := h.service.CountPresent(h.getMeetingID())
	c.JSON(http.StatusOK, gin.H{"ok": true, "present_count": count})
}

// GET /api/attendance/count
func (h *Handler) GetCount(c *gin.Context) {
	count, err := h.service.CountPresent(h.getMeetingID())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sayım alınamadı"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"present_count": count})
}

// GET /api/attendance/present
func (h *Handler) GetPresent(c *gin.Context) {
	members, err := h.service.GetPresentMembers(h.getMeetingID())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "üye listesi alınamadı"})
		return
	}
	c.JSON(http.StatusOK, members)
}
