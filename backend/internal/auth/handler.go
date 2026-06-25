package auth

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	db         *sql.DB
	attendance AttendanceChecker
	meetingID  string
}

type AttendanceChecker interface {
	IsMemberPresent(meetingID string, memberID int) (bool, error)
}

func NewHandler(db *sql.DB, attendance AttendanceChecker, meetingID string) *Handler {
	return &Handler{db: db, attendance: attendance, meetingID: meetingID}
}

type loginRequest struct {
	TCNO     string `json:"tc_no" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Member struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// POST /api/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tc_no ve password zorunlu"})
		return
	}

	var member Member
	var passwordHash string
	query := `SELECT id, name, role, password_hash FROM members WHERE tc_no = $1`
	err := h.db.QueryRow(query, req.TCNO).Scan(
		&member.ID, &member.Name, &member.Role, &passwordHash,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "geçersiz kimlik bilgileri"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sunucu hatası"})
		return
	}

	// Bcrypt şifre doğrulama
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "geçersiz kimlik bilgileri"})
		return
	}

	// Salonda mı kontrol et (Moderatör muaf)
	if member.Role != "moderator" {
		present, err := h.attendance.IsMemberPresent(h.meetingID, member.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "yoklama kontrol hatası"})
			return
		}
		if !present {
			c.JSON(http.StatusForbidden, gin.H{"error": "salonda değilsiniz"})
			return
		}
	}

	c.SetCookie(
		"session_member_id",
		strconv.Itoa(member.ID),
		3600*8,
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"ok":     true,
		"member": member,
	})
}

// POST /api/auth/logout
func (h *Handler) Logout(c *gin.Context) {
	c.SetCookie("session_member_id", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GET /api/auth/me
func (h *Handler) Me(c *gin.Context) {
	memberID, exists := c.Get("member_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "oturum yok"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"member_id": memberID})
}
