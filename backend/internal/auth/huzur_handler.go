package auth

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HuzurHandler struct {
	db           *sql.DB
	getMeetingID func() string
}

func NewHuzurHandler(db *sql.DB, getMeetingID func() string) *HuzurHandler {
	return &HuzurHandler{db: db, getMeetingID: getMeetingID}
}

type HuzurRow struct {
	MemberID   int     `json:"member_id"`
	Name       string  `json:"name"`
	Soyisim    string  `json:"soyisim"`
	RoleName   string  `json:"role_name"`
	Katsayi    float64 `json:"katsayi"`
	TabanTutar float64 `json:"taban_tutar"`
	Hakkedis   float64 `json:"hakkedis"`
	Katildi    bool    `json:"katildi"`
}

// GET /api/huzur/list
func (h *HuzurHandler) List(c *gin.Context) {
	// Taban tutarı al
	var tabanTutar float64
	h.db.QueryRow(
		`SELECT taban_tutar FROM huzur_hakki_settings WHERE meeting_id = $1`,
		h.getMeetingID(),
	).Scan(&tabanTutar)

	rows, err := h.db.Query(`
		SELECT 
			m.id,
			m.name,
			COALESCE(m.soyisim, ''),
			COALESCE(mr.role_name, 'Üye'),
			COALESCE(mr.katsayi, 1.0),
			EXISTS (
				SELECT 1 FROM attendance_sessions a
				WHERE a.member_id = m.id 
				AND a.meeting_id = $1
				AND a.exited_at IS NULL
			) as katildi
		FROM members m
		LEFT JOIN member_roles mr ON mr.id = m.member_role_id
		WHERE m.role = 'member'
		ORDER BY mr.katsayi DESC, m.name
	`, h.getMeetingID())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "liste alınamadı"})
		return
	}
	defer rows.Close()

	var list []HuzurRow
	for rows.Next() {
		var r HuzurRow
		rows.Scan(&r.MemberID, &r.Name, &r.Soyisim, &r.RoleName, &r.Katsayi, &r.Katildi)
		r.TabanTutar = tabanTutar
		if r.Katildi {
			r.Hakkedis = tabanTutar * r.Katsayi
		}
		list = append(list, r)
	}
	if list == nil {
		list = []HuzurRow{}
	}
	c.JSON(http.StatusOK, gin.H{
		"taban_tutar": tabanTutar,
		"list":        list,
	})
}

// POST /api/huzur/settings
func (h *HuzurHandler) SaveSettings(c *gin.Context) {
	var req struct {
		TabanTutar float64 `json:"taban_tutar" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.db.Exec(`
		INSERT INTO huzur_hakki_settings (meeting_id, taban_tutar)
		VALUES ($1, $2)
		ON CONFLICT (meeting_id) DO UPDATE SET taban_tutar = $2
	`, h.getMeetingID(), req.TabanTutar)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "kayıt başarısız"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GET /api/huzur/roles
func (h *HuzurHandler) GetRoles(c *gin.Context) {
	rows, err := h.db.Query(`SELECT id, role_name, katsayi FROM member_roles ORDER BY katsayi DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "roller alınamadı"})
		return
	}
	defer rows.Close()

	type Role struct {
		ID       int     `json:"id"`
		RoleName string  `json:"role_name"`
		Katsayi  float64 `json:"katsayi"`
	}
	var list []Role
	for rows.Next() {
		var r Role
		rows.Scan(&r.ID, &r.RoleName, &r.Katsayi)
		list = append(list, r)
	}
	if list == nil {
		list = []Role{}
	}
	c.JSON(http.StatusOK, list)
}

// POST /api/huzur/member-role
func (h *HuzurHandler) SetMemberRole(c *gin.Context) {
	var req struct {
		MemberID int `json:"member_id" binding:"required"`
		RoleID   int `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.db.Exec(
		`UPDATE members SET member_role_id = $2 WHERE id = $1`,
		req.MemberID, req.RoleID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rol güncellenemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
