package auth

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type MembersHandler struct {
	db    *sql.DB
	audit *AuditService
}

func NewMembersHandler(db *sql.DB, audit *AuditService) *MembersHandler {
	return &MembersHandler{db: db, audit: audit}
}

type MemberRow struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Soyisim      string `json:"soyisim"`
	TCNO         string `json:"tc_no"`
	Parti        string `json:"parti"`
	Role         string `json:"role"`
	MemberRoleID *int   `json:"member_role_id"`
}

// GET /api/members
func (h *MembersHandler) List(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, name, COALESCE(soyisim,''), tc_no, COALESCE(parti,''), role, member_role_id
		FROM members ORDER BY id
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "üyeler alınamadı"})
		return
	}
	defer rows.Close()

	var members []MemberRow
	for rows.Next() {
		var m MemberRow
		rows.Scan(&m.ID, &m.Name, &m.Soyisim, &m.TCNO, &m.Parti, &m.Role, &m.MemberRoleID)
		members = append(members, m)
	}
	if members == nil {
		members = []MemberRow{}
	}
	c.JSON(http.StatusOK, members)
}

// POST /api/members
func (h *MembersHandler) Create(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Soyisim  string `json:"soyisim"`
		TCNO     string `json:"tc_no" binding:"required"`
		Parti    string `json:"parti"`
		Role     string `json:"role"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "şifre hashlenemedi"})
		return
	}

	role := req.Role
	if role == "" {
		role = "member"
	}

	var id int
	err = h.db.QueryRow(`
		INSERT INTO members (name, soyisim, tc_no, parti, role, password_hash)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, req.Name, req.Soyisim, req.TCNO, req.Parti, role, string(hash)).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "üye eklenemedi"})
		return
	}
	h.audit.Log(AuditEntry{
		Action:     "member_create",
		TargetType: "member",
		TargetID:   strconv.Itoa(id),
		Detail:     req.Name + " " + req.Soyisim + " (" + req.TCNO + ")",
		IP:         c.ClientIP(),
	})
	c.JSON(http.StatusOK, gin.H{"ok": true, "id": id})
}

// PUT /api/members/:id
func (h *MembersHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name    string `json:"name" binding:"required"`
		Soyisim string `json:"soyisim"`
		TCNO    string `json:"tc_no" binding:"required"`
		Parti   string `json:"parti"`
		Role    string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.db.Exec(`
		UPDATE members SET name=$1, soyisim=$2, tc_no=$3, parti=$4, role=$5
		WHERE id=$6
	`, req.Name, req.Soyisim, req.TCNO, req.Parti, req.Role, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "güncelleme başarısız"})
		return
	}
	h.audit.Log(AuditEntry{
		Action:     "member_update",
		TargetType: "member",
		TargetID:   id,
		Detail:     req.Name + " " + req.Soyisim,
		IP:         c.ClientIP(),
	})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DELETE /api/members/:id
func (h *MembersHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	// Önce ilişkili kayıtları sil
	h.db.Exec(`UPDATE seats SET member_id = NULL WHERE member_id=$1`, id)
	h.db.Exec(`DELETE FROM votes WHERE member_id=$1`, id)
	h.db.Exec(`DELETE FROM attendance_sessions WHERE member_id=$1`, id)

	_, err := h.db.Exec(`DELETE FROM members WHERE id=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "silme başarısız: " + err.Error()})
		return
	}
	h.audit.Log(AuditEntry{
		Action:     "member_delete",
		TargetType: "member",
		TargetID:   id,
		IP:         c.ClientIP(),
	})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /api/members/:id/reset-password
func (h *MembersHandler) ResetPassword(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "şifre hashlenemedi"})
		return
	}

	_, err = h.db.Exec(`UPDATE members SET password_hash=$1 WHERE id=$2`, string(hash), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "şifre güncellenemedi"})
		return
	}
	h.audit.Log(AuditEntry{
		Action:     "password_reset",
		TargetType: "member",
		TargetID:   id,
		IP:         c.ClientIP(),
	})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
