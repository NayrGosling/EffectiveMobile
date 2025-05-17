package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

// NewDB создает новое соединение с базой данных PostgreSQL.
// Принимает строку подключения connStr.
// Возвращает указатель на объект sql.DB и ошибку, если она возникла.
func NewDB(connStr string) (*sql.DB, error) {
	// Создаем новое соединение с базой данных PostgreSQL
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		// Возвращаем nil и ошибку, если не удалось создать соединение
		return nil, err
	}

	// Проверяем соединение с базой данных
	if err := db.Ping(); err != nil {
		// Возвращаем nil и ошибку, если соединение не удалось установить
		return nil, err
	}

	// Возвращаем указатель на объект sql.DB и nil, если ошибок не возникло
	return db, nil
}
