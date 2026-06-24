//go:build ignore

package main

import (
	"log"
	"os"
	"strings"

	"meclis-oylama/backend/internal/db"
)

func main() {
	if data, err := os.ReadFile(".env"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}
	}

	dbURL := os.Getenv("DATABASE_URL")
	database, err := db.Connect(dbURL)
	if err != nil {
		log.Fatalf("Bağlantı hatası: %v", err)
	}
	defer database.Close()

	if err := db.RunMigrations(database, "internal/db/migrations"); err != nil {
		log.Fatalf("Migration hatası: %v", err)
	}

	log.Println("Tüm migration'lar tamamlandı.")
}
