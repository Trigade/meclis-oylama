package attendance

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) RecordEntry(meetingID string, memberID int) error {
	query := `
		INSERT INTO attendance_sessions (meeting_id, member_id, entered_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (meeting_id, member_id)
		DO UPDATE SET entered_at = $3, exited_at = NULL
	`
	_, err := s.db.Exec(query, meetingID, memberID, time.Now())
	if err != nil {
		return fmt.Errorf("giriş kaydedilemedi: %w", err)
	}
	log.Printf("Giriş kaydedildi → toplantı: %s, üye: %d", meetingID, memberID)
	return nil
}

func (s *Service) RecordExit(meetingID string, memberID int) error {
	query := `
		UPDATE attendance_sessions
		SET exited_at = $3
		WHERE meeting_id = $1 AND member_id = $2 AND exited_at IS NULL
	`
	_, err := s.db.Exec(query, meetingID, memberID, time.Now())
	if err != nil {
		return fmt.Errorf("çıkış kaydedilemedi: %w", err)
	}
	log.Printf("Çıkış kaydedildi → toplantı: %s, üye: %d", meetingID, memberID)
	return nil
}

func (s *Service) CountPresent(meetingID string) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM attendance_sessions
		WHERE meeting_id = $1 AND exited_at IS NULL
	`
	err := s.db.QueryRow(query, meetingID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("üye sayısı alınamadı: %w", err)
	}
	return count, nil
}

func (s *Service) IsMemberPresent(meetingID string, memberID int) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM attendance_sessions
			WHERE meeting_id = $1 AND member_id = $2 AND exited_at IS NULL
		)
	`
	err := s.db.QueryRow(query, meetingID, memberID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("üye varlığı kontrol edilemedi: %w", err)
	}
	return exists, nil
}
