package auth

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	db *sql.DB
}

func NewAuditHandler(db *sql.DB) *AuditHandler {
	return &AuditHandler{db: db}
}

type AuditRow struct {
	ID         int    `json:"id"`
	ActorName  string `json:"actor_name"`
	Action     string `json:"action"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Detail     string `json:"detail"`
	IP         string `json:"ip"`
	CreatedAt  string `json:"created_at"`
}

// GET /api/audit
func (h *AuditHandler) List(c *gin.Context) {
	limit := "100"
	if l := c.Query("limit"); l != "" {
		limit = l
	}

	action := c.Query("action")
	var rows *sql.Rows
	var err error

	if action != "" {
		rows, err = h.db.Query(`
			SELECT id, COALESCE(actor_name,''), action, COALESCE(target_type,''),
			       COALESCE(target_id,''), COALESCE(detail,''), COALESCE(ip,''), created_at::text
			FROM audit_logs WHERE action = $1
			ORDER BY created_at DESC LIMIT $2
		`, action, limit)
	} else {
		rows, err = h.db.Query(`
			SELECT id, COALESCE(actor_name,''), action, COALESCE(target_type,''),
			       COALESCE(target_id,''), COALESCE(detail,''), COALESCE(ip,''), created_at::text
			FROM audit_logs
			ORDER BY created_at DESC LIMIT $1
		`, limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "loglar alınamadı"})
		return
	}
	defer rows.Close()

	var list []AuditRow
	for rows.Next() {
		var r AuditRow
		rows.Scan(&r.ID, &r.ActorName, &r.Action, &r.TargetType,
			&r.TargetID, &r.Detail, &r.IP, &r.CreatedAt)
		list = append(list, r)
	}
	if list == nil {
		list = []AuditRow{}
	}
	c.JSON(http.StatusOK, list)
}
