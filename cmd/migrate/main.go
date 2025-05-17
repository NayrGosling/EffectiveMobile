package main

import (
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"

	"effect/internal/config"
	"effect/internal/db"
	"effect/internal/migrations"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	log.SetLevel(cfg.LogLevel)
	log.Info("running migrations")

	dbConn, err := db.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to DB: %v", err)
	}
	log.Debug("connected to database for migrations")

	os.Setenv("MIGRATIONS_DIR", cfg.MigrationsDir)
	if err := migrations.Run(dbConn); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}
	log.Info("migrations applied successfully")
}
