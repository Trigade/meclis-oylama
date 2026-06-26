package auth

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SalonHandler struct {
	db           *sql.DB
	getMeetingID func() string
}

func NewSalonHandler(db *sql.DB, getMeetingID func() string) *SalonHandler {
	return &SalonHandler{db: db, getMeetingID: getMeetingID}
}

type SeatRow struct {
	SeatNo    int     `json:"seat_no"`
	BlockName string  `json:"block_name"`
	MemberID  *int    `json:"member_id"`
	Name      *string `json:"name"`
	Soyisim   *string `json:"soyisim"`
	Present   bool    `json:"present"`
}

// GET /api/salon/seats
func (h *SalonHandler) GetSeats(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT 
			s.seat_no,
			COALESCE(s.block_name, ''),
			s.member_id,
			m.name,
			m.soyisim,
			EXISTS (
				SELECT 1 FROM attendance_sessions a
				WHERE a.member_id = s.member_id
				AND a.meeting_id = $1
				AND a.exited_at IS NULL
			) as present
		FROM seats s
		LEFT JOIN members m ON m.id = s.member_id
		ORDER BY s.seat_no
	`, h.getMeetingID())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "koltuklar alınamadı"})
		return
	}
	defer rows.Close()

	var list []SeatRow
	for rows.Next() {
		var r SeatRow
		rows.Scan(&r.SeatNo, &r.BlockName, &r.MemberID, &r.Name, &r.Soyisim, &r.Present)
		list = append(list, r)
	}
	if list == nil {
		list = []SeatRow{}
	}
	c.JSON(http.StatusOK, list)
}
