package main

import (
	"log"
	"os"

	"effect/internal/config"
	"effect/internal/db"
	"effect/internal/migrations"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = config.GetDBConnString()
	}

	dbConn, err := db.NewDB(dsn)
	if err != nil {
		log.Fatalf("connect to DB: %v", err)
	}

	if err := migrations.Run(dbConn); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}
	log.Println("migrations applied")
}
