package main

import (
	"log"
	"os"

	"effect/internal/config"
	"effect/internal/db"
	"effect/internal/migrations"
)

func main() {
	// Получаем строку подключения
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = config.GetDBConnString()
	}

	// Подключаемся к БД
	database, err := db.NewDB(connStr)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}

	// Запускаем миграции
	if err := migrations.Run(database); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	log.Println("migrations applied successfully")
}
