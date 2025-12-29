package repository

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// OpenDB открывает (или создаёт) SQLite БД
func OpenDB(path string) (*sql.DB, error) {
	// Убеждаемся, что директория существует
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Включаем поддержку внешних ключей
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return nil, err
	}

	return db, nil
}
