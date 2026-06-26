package auth

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type KomisyonHandler struct {
	db    *sql.DB
	audit *AuditService
}

func NewKomisyonHandler(db *sql.DB, audit *AuditService) *KomisyonHandler {
	return &KomisyonHandler{db: db, audit: audit}
}

type Komisyon struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Aciklama  string  `json:"aciklama"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
	EndedAt   *string `json:"ended_at"`
	UyeSayisi int     `json:"uye_sayisi"`
}

// GET /api/komisyonlar
func (h *KomisyonHandler) List(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT k.id, k.name, COALESCE(k.aciklama,''), k.status, k.created_at::text,
		       k.ended_at::text,
		       COUNT(ku.member_id) as uye_sayisi
		FROM komisyonlar k
		LEFT JOIN komisyon_uyeler ku ON ku.komisyon_id = k.id
		GROUP BY k.id
		ORDER BY k.id DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "komisyonlar alınamadı"})
		return
	}
	defer rows.Close()

	var list []Komisyon
	for rows.Next() {
		var k Komisyon
		rows.Scan(&k.ID, &k.Name, &k.Aciklama, &k.Status,
			&k.CreatedAt, &k.EndedAt, &k.UyeSayisi)
		list = append(list, k)
	}
	if list == nil {
		list = []Komisyon{}
	}
	c.JSON(http.StatusOK, list)
}

// POST /api/komisyonlar
func (h *KomisyonHandler) Create(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Aciklama string `json:"aciklama"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var k Komisyon
	err := h.db.QueryRow(`
		INSERT INTO komisyonlar (name, aciklama)
		VALUES ($1, $2)
		RETURNING id, name, COALESCE(aciklama,''), status, created_at::text
	`, req.Name, req.Aciklama).Scan(&k.ID, &k.Name, &k.Aciklama, &k.Status, &k.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "komisyon oluşturulamadı"})
		return
	}

	h.audit.Log(AuditEntry{
		Action:     "komisyon_create",
		TargetType: "komisyon",
		Detail:     req.Name,
		IP:         c.ClientIP(),
	})

	c.JSON(http.StatusOK, k)
}

// POST /api/komisyonlar/:id/end
func (h *KomisyonHandler) End(c *gin.Context) {
	id := c.Param("id")
	_, err := h.db.Exec(`
		UPDATE komisyonlar SET status='ended', ended_at=NOW() WHERE id=$1
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "komisyon kapatılamadı"})
		return
	}
	h.audit.Log(AuditEntry{
		Action:     "komisyon_end",
		TargetType: "komisyon",
		TargetID:   id,
		IP:         c.ClientIP(),
	})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GET /api/komisyonlar/:id/uyeler
func (h *KomisyonHandler) GetUyeler(c *gin.Context) {
	id := c.Param("id")
	rows, err := h.db.Query(`
		SELECT m.id, m.name, COALESCE(m.soyisim,''), COALESCE(m.parti,'')
		FROM komisyon_uyeler ku
		JOIN members m ON m.id = ku.member_id
		WHERE ku.komisyon_id = $1
		ORDER BY m.name
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "üyeler alınamadı"})
		return
	}
	defer rows.Close()

	type KUye struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Soyisim string `json:"soyisim"`
		Parti   string `json:"parti"`
	}
	var list []KUye
	for rows.Next() {
		var u KUye
		rows.Scan(&u.ID, &u.Name, &u.Soyisim, &u.Parti)
		list = append(list, u)
	}
	if list == nil {
		list = []KUye{}
	}
	c.JSON(http.StatusOK, list)
}

// POST /api/komisyonlar/:id/uyeler
func (h *KomisyonHandler) AddUye(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		MemberID int `json:"member_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.db.Exec(`
		INSERT INTO komisyon_uyeler (komisyon_id, member_id)
		VALUES ($1, $2) ON CONFLICT DO NOTHING
	`, id, req.MemberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "üye eklenemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DELETE /api/komisyonlar/:id/uyeler/:mid
func (h *KomisyonHandler) RemoveUye(c *gin.Context) {
	id := c.Param("id")
	mid := c.Param("mid")
	h.db.Exec(`DELETE FROM komisyon_uyeler WHERE komisyon_id=$1 AND member_id=$2`, id, mid)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GET /api/komisyonlar/:id/kararlar
func (h *KomisyonHandler) GetKararlar(c *gin.Context) {
	id := c.Param("id")
	rows, err := h.db.Query(`
		SELECT id, karar_metni, created_at::text
		FROM komisyon_kararlar WHERE komisyon_id=$1
		ORDER BY id DESC
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "kararlar alınamadı"})
		return
	}
	defer rows.Close()

	type Karar struct {
		ID         int    `json:"id"`
		KararMetni string `json:"karar_metni"`
		CreatedAt  string `json:"created_at"`
	}
	var list []Karar
	for rows.Next() {
		var k Karar
		rows.Scan(&k.ID, &k.KararMetni, &k.CreatedAt)
		list = append(list, k)
	}
	if list == nil {
		list = []Karar{}
	}
	c.JSON(http.StatusOK, list)
}

// POST /api/komisyonlar/:id/kararlar
func (h *KomisyonHandler) AddKarar(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		KararMetni string `json:"karar_metni" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var kid int
	err := h.db.QueryRow(`
		INSERT INTO komisyon_kararlar (komisyon_id, karar_metni)
		VALUES ($1, $2) RETURNING id
	`, id, req.KararMetni).Scan(&kid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "karar eklenemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "id": kid})
}

// DELETE /api/komisyonlar/:id/kararlar/:kid
func (h *KomisyonHandler) DeleteKarar(c *gin.Context) {
	kid := c.Param("kid")
	h.db.Exec(`DELETE FROM komisyon_kararlar WHERE id=$1`, kid)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
