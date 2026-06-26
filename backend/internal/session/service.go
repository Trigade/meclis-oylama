package session

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
)

// ActiveSession tüm handler'lar tarafından paylaşılan aktif oturum bilgisi
type ActiveSession struct {
	mu        sync.RWMutex
	MeetingID string
}

func NewActiveSession() *ActiveSession {
	return &ActiveSession{MeetingID: "meeting-default"}
}

func (s *ActiveSession) Get() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.MeetingID
}

func (s *ActiveSession) Set(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.MeetingID = id
	log.Printf("Aktif meetingID güncellendi → %s", id)
}

// LoadFromDB — başlangıçta aktif oturumu DB'den yükler
func (s *ActiveSession) LoadFromDB(db *sql.DB) error {
	var id int
	err := db.QueryRow(`SELECT id FROM meetings WHERE status = 'active' LIMIT 1`).Scan(&id)
	if err == sql.ErrNoRows {
		log.Println("Aktif oturum bulunamadı, varsayılan meetingID kullanılıyor.")
		return nil
	}
	if err != nil {
		return fmt.Errorf("aktif oturum sorgulanamadı: %w", err)
	}
	s.Set(fmt.Sprintf("meeting-%d", id))
	return nil
}
