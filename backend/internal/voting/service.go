package voting

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"
)

const ToplamUye = 32

type EsikTipi string

const (
	Oybirligi    EsikTipi = "oybirligi"
	SaltCogunluk EsikTipi = "salt_cogunluk"
	IkideUc      EsikTipi = "iki_uc"
	UcteUc       EsikTipi = "uc_dort"
)

func hesaplaEsik(esik EsikTipi, toplam int) int {
	switch esik {
	case Oybirligi:
		return toplam
	case IkideUc:
		return int(math.Ceil(float64(toplam) * 2 / 3))
	case UcteUc:
		return int(math.Ceil(float64(toplam) * 3 / 4))
	default: // SaltCogunluk
		return toplam/2 + 1
	}
}

type Service struct {
	db  *sql.DB
	hub *Hub
}

func NewService(db *sql.DB, hub *Hub) *Service {
	return &Service{db: db, hub: hub}
}

type Voting struct {
	ID         int        `json:"id"`
	MeetingID  string     `json:"meeting_id"`
	Title      string     `json:"title"`
	Status     string     `json:"status"`
	OylamaTipi string     `json:"oylama_tipi"`
	EsikTipi   string     `json:"esik_tipi"`
	EsikSayi   int        `json:"esik_sayi"`
	Result     *string    `json:"result,omitempty"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
}

type StartParams struct {
	MeetingID    string
	Title        string
	OylamaTipi   string // "gizli" | "acik"
	EsikTipi     EsikTipi
	PresentCount int
}

func (s *Service) Start(p StartParams) (*Voting, error) {
	if p.PresentCount < 16 {
		return nil, fmt.Errorf("yetersiz üye: %d (minimum 16)", p.PresentCount)
	}

	esikSayi := hesaplaEsik(p.EsikTipi, ToplamUye)

	now := time.Now()
	var v Voting
	query := `
		INSERT INTO votings (meeting_id, title, status, oylama_tipi, esik_tipi, esik_sayi, toplam_uye, started_at)
		VALUES ($1, $2, 'active', $3, $4, $5, $6, $7)
		RETURNING id, meeting_id, title, status, oylama_tipi, esik_tipi, esik_sayi, started_at
	`
	err := s.db.QueryRow(query,
		p.MeetingID, p.Title, p.OylamaTipi, string(p.EsikTipi), esikSayi, ToplamUye, now,
	).Scan(&v.ID, &v.MeetingID, &v.Title, &v.Status, &v.OylamaTipi, &v.EsikTipi, &v.EsikSayi, &v.StartedAt)
	if err != nil {
		return nil, fmt.Errorf("oylama oluşturulamadı: %w", err)
	}

	msg, _ := json.Marshal(map[string]any{
		"type":        "voting_started",
		"voting_id":   v.ID,
		"title":       v.Title,
		"oylama_tipi": v.OylamaTipi,
		"esik_sayi":   v.EsikSayi,
		"started_at":  v.StartedAt,
	})
	s.hub.Broadcast(msg)
	log.Printf("Oylama başlatıldı → ID: %d, tip: %s, eşik: %d", v.ID, p.OylamaTipi, esikSayi)

	return &v, nil
}

func (s *Service) CastVote(votingID, memberID int, choice string) error {
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

	s.broadcastCounts(votingID)
	return nil
}

func (s *Service) Finalize(votingID int) (*Voting, error) {
	var esikSayi int
	var oylamaTipi string
	s.db.QueryRow(
		`SELECT esik_sayi, oylama_tipi FROM votings WHERE id = $1`, votingID,
	).Scan(&esikSayi, &oylamaTipi)

	var yesCount int
	s.db.QueryRow(
		`SELECT COUNT(*) FROM votes WHERE voting_id = $1 AND choice = 'evet'`, votingID,
	).Scan(&yesCount)

	result := "reddedildi"
	if yesCount >= esikSayi {
		result = "kabul_edildi"
	}

	var v Voting
	query := `
		UPDATE votings SET status = 'closed', result = $2, ended_at = NOW()
		WHERE id = $1
		RETURNING id, meeting_id, title, status, oylama_tipi, esik_tipi, esik_sayi, result, started_at
	`
	err := s.db.QueryRow(query, votingID, result).Scan(
		&v.ID, &v.MeetingID, &v.Title, &v.Status, &v.OylamaTipi, &v.EsikTipi, &v.EsikSayi, &v.Result, &v.StartedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("oylama kapatılamadı: %w", err)
	}

	msg, _ := json.Marshal(map[string]any{
		"type":      "voting_closed",
		"voting_id": v.ID,
		"result":    result,
		"yes_count": yesCount,
		"esik_sayi": esikSayi,
	})
	s.hub.Broadcast(msg)
	log.Printf("Oylama kapandı → ID: %d, sonuç: %s (%d/%d evet)", v.ID, result, yesCount, esikSayi)

	return &v, nil
}

func (s *Service) broadcastCounts(votingID int) {
	var yes, no int
	var oylamaTipi string
	s.db.QueryRow(`SELECT COUNT(*) FROM votes WHERE voting_id=$1 AND choice='evet'`, votingID).Scan(&yes)
	s.db.QueryRow(`SELECT COUNT(*) FROM votes WHERE voting_id=$1 AND choice='hayir'`, votingID).Scan(&no)
	s.db.QueryRow(`SELECT oylama_tipi FROM votings WHERE id=$1`, votingID).Scan(&oylamaTipi)

	msg, _ := json.Marshal(map[string]any{
		"type":        "vote_update",
		"voting_id":   votingID,
		"yes":         yes,
		"no":          no,
		"oylama_tipi": oylamaTipi,
	})
	s.hub.Broadcast(msg)
}
