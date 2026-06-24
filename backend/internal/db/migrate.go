package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
)

func RunMigrations(db *sql.DB, migrationsDir string) error {
	// migrations tablosunu oluştur
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("migrations tablosu oluşturulamadı: %w", err)
	}

	// SQL dosyalarını bul
	pattern := filepath.Join(migrationsDir, "*.sql")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("migration dosyaları okunamadı: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		filename := filepath.Base(file)

		// Daha önce çalıştırıldı mı?
		var count int
		db.QueryRow(
			`SELECT COUNT(*) FROM schema_migrations WHERE filename = $1`, filename,
		).Scan(&count)
		if count > 0 {
			log.Printf("Migration atlandı (zaten uygulandı): %s", filename)
			continue
		}

		// Dosyayı oku ve çalıştır
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("%s okunamadı: %w", filename, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("%s çalıştırılamadı: %w", filename, err)
		}

		db.Exec(
			`INSERT INTO schema_migrations (filename) VALUES ($1)`, filename,
		)
		log.Printf("Migration uygulandı: %s", filename)
	}

	return nil
}
