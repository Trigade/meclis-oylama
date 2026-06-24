package voting

import (
	"net/http"
	"strconv"

	"meclis-oylama/backend/internal/attendance"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service    *Service
	hub        *Hub
	attendance *attendance.Service
	meetingID  string
}

func NewHandler(service *Service, hub *Hub, attendance *attendance.Service, meetingID string) *Handler {
	return &Handler{
		service:    service,
		hub:        hub,
		attendance: attendance,
		meetingID:  meetingID,
	}
}

type startRequest struct {
	Title string `json:"title" binding:"required"`
}

type voteRequest struct {
	Choice string `json:"choice" binding:"required,oneof=evet hayir"`
}

// POST /api/voting/start  (moderatör)
func (h *Handler) Start(c *gin.Context) {
	var req startRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	count, err := h.attendance.CountPresent(h.meetingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "üye sayısı alınamadı"})
		return
	}

	voting, err := h.service.Start(h.meetingID, req.Title, count)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, voting)
}

// POST /api/voting/:id/vote  (üye)
func (h *Handler) CastVote(c *gin.Context) {
	votingID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "geçersiz oylama ID"})
		return
	}

	memberID := c.GetInt("member_id")

	var req voteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CastVote(votingID, memberID, req.Choice); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /api/voting/:id/finalize  (moderatör)
func (h *Handler) Finalize(c *gin.Context) {
	votingID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "geçersiz oylama ID"})
		return
	}

	result, err := h.service.Finalize(votingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GET /api/voting/ws  (WebSocket)
func (h *Handler) WS(c *gin.Context) {
	h.hub.ServeWS(c)
}
