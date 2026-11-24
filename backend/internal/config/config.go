package config

// Пакет config — одно место, где живут настройки приложения.
// Все остальные файлы берут настройки отсюда, а не хардкодят их у себя.

import (
	"fmt" // чтобы собрать строку через Sprintf
	"os"  // чтобы читать переменные окружения
)

// Config — структура с настройками.
// Два поля: куда класть базу и на каком порту слушать.
type Config struct {
	DBPath string // путь к файлу базы данных
	Port   string // номер порта, например "8080"
}

// Load читает настройки из переменных окружения.
// Если переменная не задана — подставляет значение по умолчанию.
//
// Пример: DB_PATH=C:\db\lab1.db PORT=9090 go run ./cmd/server
// Если ничего не задать — будут умолчания: ../database/lab1.db и 8080.
func Load() Config {
	cfg := Config{
		DBPath: os.Getenv("DB_PATH"),
		Port:   os.Getenv("PORT"),
	}

	// Если DB_PATH пустая — ставим путь по умолчанию.
	// "../database/lab1.db" значит: на уровень выше папки backend, потом в database.
	if cfg.DBPath == "" {
		cfg.DBPath = "../database/lab1.db"
	}

	// Если PORT пустой — 8080.
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg
}

// DSN собирает строку подключения к SQLite.
// DSN = Data Source Name, «имя источника данных».
//
// Что внутри:
//
//	file:путь          — открыть файл базы
//	?_pragma=foreign_keys(1) — включить проверку внешних ключей.
//	    Без этого SQLite разрешит удалить должность, на которой есть сотрудники,
//	    хотя в init.sql написано ON DELETE RESTRICT.
//	    Это критично для правила «нельзя удалить занятую должность».
func (c Config) DSN() string {
	return fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", c.DBPath)
}
