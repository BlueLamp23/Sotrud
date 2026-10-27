package repository

import (
	"database/sql"
	"log"
	_ "modernc.org/sqlite" // Драйвер SQLite (чистый Go, без CGO)
)

// NewDB создает подключение к базе данных и настраивает его
func NewDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Проверка соединения
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// КРИТИЧЕСКИ ВАЖНО: включаем внешние ключи для этого соединения
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, err
	}
	
	// Включаем режим WAL для лучшей производительности при чтении/записи
	_, err = db.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		log.Println("Warning: Failed to set WAL mode:", err)
	}

	return db, nil
}