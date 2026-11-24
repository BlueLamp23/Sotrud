package repository

import (
	"database/sql"
	"log"
	_ "modernc.org/sqlite" // Драйвер SQLite
)

// NewDB создает подключение к базе данных и настраивает его.
// Внешние ключи (foreign_keys) включаются через DSN в config.DSN()
// (?_pragma=foreign_keys(1)) — на всех соединениях пула сразу.
func NewDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Проверка соединения
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// включаем режим WAL для лучшей производительности при чтении/записи
	_, err = db.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		log.Println("Warning: Failed to set WAL mode:", err)
	}

	return db, nil
}
