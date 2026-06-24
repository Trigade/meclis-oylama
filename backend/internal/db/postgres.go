package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("veritabanı açılamadı: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("veritabanına bağlanılamadı: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Println("PostgreSQL bağlantısı kuruldu.")
	return db, nil
}
