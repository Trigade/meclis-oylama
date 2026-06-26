package auth

import (
	"database/sql"
	"log"
)

type AuditService struct {
	db *sql.DB
}

func NewAuditService(db *sql.DB) *AuditService {
	return &AuditService{db: db}
}

type AuditEntry struct {
	ActorID    *int
	ActorName  string
	Action     string
	TargetType string
	TargetID   string
	Detail     string
	IP         string
}

func (s *AuditService) Log(e AuditEntry) {
	_, err := s.db.Exec(`
		INSERT INTO audit_logs (actor_id, actor_name, action, target_type, target_id, detail, ip)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, e.ActorID, e.ActorName, e.Action, e.TargetType, e.TargetID, e.Detail, e.IP)
	if err != nil {
		log.Printf("Audit log hatası: %v", err)
	}
}
