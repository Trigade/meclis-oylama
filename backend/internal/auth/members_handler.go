package auth

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MembersHandler struct {
	db *sql.DB
}

func NewMembersHandler(db *sql.DB) *MembersHandler {
	return &MembersHandler{db: db}
}

type MemberRow struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Soyisim string `json:"soyisim"`
	TCNO    string `json:"tc_no"`
	Parti   string `json:"parti"`
	Role    string `json:"role"`
}

// GET /api/members
func (h *MembersHandler) List(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, name, COALESCE(soyisim,''), tc_no, COALESCE(parti,''), role
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
		rows.Scan(&m.ID, &m.Name, &m.Soyisim, &m.TCNO, &m.Parti, &m.Role)
		members = append(members, m)
	}
	if members == nil {
		members = []MemberRow{}
	}
	c.JSON(http.StatusOK, members)
}
