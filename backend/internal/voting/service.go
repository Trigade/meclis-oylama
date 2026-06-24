package voting

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

const (
	MinPresent  = 16
	MinYesVotes = 17
)

type Service struct {
	db  *sql.DB
	hub *Hub
}

func NewService(db *sql.DB, hub *Hub) *Service {
	return &Service{db: db, hub: hub}
}

type Voting struct {
	ID        int        `json:"id"`
	MeetingID string     `json:"meeting_id"`
	Title     string     `json:"title"`
	Status    string     `json:"status"`
	Result    *string    `json:"result,omitempty"`
	StartedAt *time.Time `json:"started_at,omitempty"`
}

// Start — yeni oylama başlatır, tüm bağlı üyelere yayın yapar
func (s *Service) Start(meetingID, title string, presentCount int) (*Voting, error) {
	if presentCount < MinPresent {
		return nil, fmt.Errorf("yetersiz üye: %d (minimum %d)", presentCount, MinPresent)
	}

	now := time.Now()
	var v Voting
	query := `
		INSERT INTO votings (meeting_id, title, status, started_at)
		VALUES ($1, $2, 'active', $3)
		RETURNING id, meeting_id, title, status, started_at
	`
	err := s.db.QueryRow(query, meetingID, title, now).Scan(
		&v.ID, &v.MeetingID, &v.Title, &v.Status, &v.StartedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("oylama oluşturulamadı: %w", err)
	}

	// Tüm bağlı üyelere anlık bildirim gönder
	msg, _ := json.Marshal(map[string]any{
		"type":       "voting_started",
		"voting_id":  v.ID,
		"title":      v.Title,
		"started_at": v.StartedAt,
	})
	s.hub.Broadcast(msg)
	log.Printf("Oylama başlatıldı → ID: %d, başlık: %s", v.ID, v.Title)

	return &v, nil
}

// CastVote — üye oyunu kullanır
func (s *Service) CastVote(votingID, memberID int, choice string) error {
	// Oylama aktif mi?
	var status string
	err := s.db.QueryRow(`SELECT status FROM votings WHERE id = $1`, votingID).Scan(&status)
	if err == sql.ErrNoRows {
		return fmt.Errorf("oylama bulunamadı")
	}
	if status != "active" {
		return fmt.Errorf("oylama aktif değil")
	}

	query := `
		INSERT INTO votes (voting_id, member_id, choice)
		VALUES ($1, $2, $3)
		ON CONFLICT (voting_id, member_id) DO NOTHING
	`
	_, err = s.db.Exec(query, votingID, memberID, choice)
	if err != nil {
		return fmt.Errorf("oy kaydedilemedi: %w", err)
	}

	// Anlık oy sayısını yayınla
	s.broadcastCounts(votingID)
	return nil
}

// Finalize — oylamayı kapatır, sonucu mühürler
func (s *Service) Finalize(votingID int) (*Voting, error) {
	var yesCount int
	s.db.QueryRow(
		`SELECT COUNT(*) FROM votes WHERE voting_id = $1 AND choice = 'evet'`, votingID,
	).Scan(&yesCount)

	result := "reddedildi"
	if yesCount >= MinYesVotes {
		result = "kabul_edildi"
	}

	var v Voting
	query := `
		UPDATE votings SET status = 'closed', result = $2, ended_at = NOW()
		WHERE id = $1
		RETURNING id, meeting_id, title, status, result, started_at
	`
	err := s.db.QueryRow(query, votingID, result).Scan(
		&v.ID, &v.MeetingID, &v.Title, &v.Status, &v.Result, &v.StartedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("oylama kapatılamadı: %w", err)
	}

	msg, _ := json.Marshal(map[string]any{
		"type":      "voting_closed",
		"voting_id": v.ID,
		"result":    result,
		"yes_count": yesCount,
	})
	s.hub.Broadcast(msg)
	log.Printf("Oylama kapandı → ID: %d, sonuç: %s (%d evet)", votingID, result, yesCount)

	return &v, nil
}

func (s *Service) broadcastCounts(votingID int) {
	var yes, no int
	s.db.QueryRow(`SELECT COUNT(*) FROM votes WHERE voting_id=$1 AND choice='evet'`, votingID).Scan(&yes)
	s.db.QueryRow(`SELECT COUNT(*) FROM votes WHERE voting_id=$1 AND choice='hayir'`, votingID).Scan(&no)

	msg, _ := json.Marshal(map[string]any{
		"type":      "vote_update",
		"voting_id": votingID,
		"yes":       yes,
		"no":        no,
	})
	s.hub.Broadcast(msg)
}
