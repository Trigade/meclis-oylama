package voting

import (
	"net/http"
	"strconv"

	"meclis-oylama/backend/internal/attendance"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service      *Service
	hub          *Hub
	attendance   *attendance.Service
	getMeetingID func() string
}

func NewHandler(service *Service, hub *Hub, attendance *attendance.Service, getMeetingID func() string) *Handler {
	return &Handler{
		service:      service,
		hub:          hub,
		attendance:   attendance,
		getMeetingID: getMeetingID,
	}
}

type startRequest struct {
	Title      string `json:"title" binding:"required"`
	OylamaTipi string `json:"oylama_tipi" binding:"required,oneof=gizli acik"`
	EsikTipi   string `json:"esik_tipi" binding:"required,oneof=oybirligi salt_cogunluk iki_uc uc_dort"`
}

type voteRequest struct {
	Choice string `json:"choice" binding:"required,oneof=evet hayir"`
}

// POST /api/voting/start
func (h *Handler) Start(c *gin.Context) {
	var req startRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// HATA DÜZELTİLDİ: h.getMeetingID()() -> h.getMeetingID()
	count, err := h.attendance.CountPresent(h.getMeetingID())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "üye sayısı alınamadı"})
		return
	}

	// HATA DÜZELTİLDİ: h.getMeetingID()() -> h.getMeetingID()
	voting, err := h.service.Start(StartParams{
		MeetingID:    h.getMeetingID(),
		Title:        req.Title,
		OylamaTipi:   req.OylamaTipi,
		EsikTipi:     EsikTipi(req.EsikTipi),
		PresentCount: count,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, voting)
}

// POST /api/voting/:id/vote
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

// POST /api/voting/:id/finalize
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

// GET /api/voting/ws
func (h *Handler) WS(c *gin.Context) {
	h.hub.ServeWS(c)
}

// GET /api/voting/active
func (h *Handler) GetActive(c *gin.Context) {
	var id int
	var title, oylamaTipi, esikTipi string
	var esikSayi int

	err := h.service.db.QueryRow(`
        SELECT id, title, oylama_tipi, esik_tipi, esik_sayi
        FROM votings WHERE status = 'active' ORDER BY id DESC LIMIT 1
    `).Scan(&id, &title, &oylamaTipi, &esikTipi, &esikSayi)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"active": false})
		return
	}

	var yes, no int
	h.service.db.QueryRow(`SELECT COUNT(*) FROM votes WHERE voting_id=$1 AND choice='evet'`, id).Scan(&yes)
	h.service.db.QueryRow(`SELECT COUNT(*) FROM votes WHERE voting_id=$1 AND choice='hayir'`, id).Scan(&no)

	c.JSON(http.StatusOK, gin.H{
		"active":      true,
		"id":          id,
		"title":       title,
		"oylama_tipi": oylamaTipi,
		"esik_tipi":   esikTipi,
		"esik_sayi":   esikSayi,
		"yes":         yes,
		"no":          no,
	})
}

// GET /api/voting/recent
func (h *Handler) GetRecent(c *gin.Context) {
	rows, err := h.service.db.Query(`
        SELECT id, title, oylama_tipi, esik_tipi, esik_sayi, status, result, started_at, ended_at
        FROM votings ORDER BY id DESC LIMIT 10
    `)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sorgu hatası"})
		return
	}
	defer rows.Close()

	type Row struct {
		ID         int     `json:"id"`
		Title      string  `json:"title"`
		OylamaTipi string  `json:"oylama_tipi"`
		EsikTipi   string  `json:"esik_tipi"`
		EsikSayi   int     `json:"esik_sayi"`
		Status     string  `json:"status"`
		Result     *string `json:"result"`
		StartedAt  *string `json:"started_at"`
		EndedAt    *string `json:"ended_at"`
	}

	var list []Row
	for rows.Next() {
		var r Row
		rows.Scan(&r.ID, &r.Title, &r.OylamaTipi, &r.EsikTipi, &r.EsikSayi, &r.Status, &r.Result, &r.StartedAt, &r.EndedAt)
		list = append(list, r)
	}
	if list == nil {
		list = []Row{}
	}
	c.JSON(http.StatusOK, list)
}

// GET /api/reports/votings
func (h *Handler) GetReport(c *gin.Context) {
	rows, err := h.service.db.Query(`
        SELECT 
            v.id, v.title, v.oylama_tipi, v.esik_tipi, v.esik_sayi,
            v.status, v.result, v.started_at, v.ended_at,
            COUNT(CASE WHEN vo.choice = 'evet' THEN 1 END) as yes_count,
            COUNT(CASE WHEN vo.choice = 'hayir' THEN 1 END) as no_count
        FROM votings v
        LEFT JOIN votes vo ON vo.voting_id = v.id
        GROUP BY v.id
        ORDER BY v.id DESC
    `)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "raporlar alınamadı"})
		return
	}
	defer rows.Close()

	type VotingReport struct {
		ID         int     `json:"id"`
		Title      string  `json:"title"`
		OylamaTipi string  `json:"oylama_tipi"`
		EsikTipi   string  `json:"esik_tipi"`
		EsikSayi   int     `json:"esik_sayi"`
		Status     string  `json:"status"`
		Result     *string `json:"result"`
		StartedAt  *string `json:"started_at"`
		EndedAt    *string `json:"ended_at"`
		YesCount   int     `json:"yes_count"`
		NoCount    int     `json:"no_count"`
	}

	var list []VotingReport
	for rows.Next() {
		var r VotingReport
		rows.Scan(&r.ID, &r.Title, &r.OylamaTipi, &r.EsikTipi, &r.EsikSayi,
			&r.Status, &r.Result, &r.StartedAt, &r.EndedAt, &r.YesCount, &r.NoCount)
		list = append(list, r)
	}
	if list == nil {
		list = []VotingReport{}
	}
	c.JSON(http.StatusOK, list)
}

// GET /api/reports/votings/:id/detail
func (h *Handler) GetVoteDetail(c *gin.Context) {
	votingID := c.Param("id")

	// Önce oylama tipini kontrol et
	var oylamaTipi string
	err := h.service.db.QueryRow(
		`SELECT oylama_tipi FROM votings WHERE id = $1`, votingID,
	).Scan(&oylamaTipi)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "oylama bulunamadı"})
		return
	}

	// Gizli oylamada detay gösterme
	if oylamaTipi == "gizli" {
		c.JSON(http.StatusForbidden, gin.H{"error": "gizli oylamada detay görüntülenemez"})
		return
	}

	rows, err := h.service.db.Query(`
        SELECT m.id, m.name, COALESCE(m.soyisim,''), COALESCE(m.parti,''), vo.choice, vo.cast_at
        FROM votes vo
        JOIN members m ON m.id = vo.member_id
        WHERE vo.voting_id = $1
        ORDER BY m.name
    `, votingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "detay alınamadı"})
		return
	}
	defer rows.Close()

	type VoteDetail struct {
		MemberID int    `json:"member_id"`
		Name     string `json:"name"`
		Soyisim  string `json:"soyisim"`
		Parti    string `json:"parti"`
		Choice   string `json:"choice"`
		CastAt   string `json:"cast_at"`
	}

	var list []VoteDetail
	for rows.Next() {
		var d VoteDetail
		rows.Scan(&d.MemberID, &d.Name, &d.Soyisim, &d.Parti, &d.Choice, &d.CastAt)
		list = append(list, d)
	}
	if list == nil {
		list = []VoteDetail{}
	}
	c.JSON(http.StatusOK, list)
}
