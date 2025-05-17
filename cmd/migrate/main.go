// cmd/migrate/main.go
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"effect/internal/config"
	"effect/internal/db"

	_ "github.com/lib/pq"
)

func main() {
	mode := flag.String("mode", "migrate", "mode to run: migrate or rollback")
	flag.Parse()

	// Загрузка конфига из env
	cfg := config.Load()
	log.Printf("Connecting to %s", cfg.DatabaseURL)

	conn, err := db.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer conn.Close()

	switch *mode {
	case "migrate":
		if err := runMigrations(conn, cfg.MigrationsDir); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
		fmt.Println("Migrations applied successfully.")
	case "rollback":
		if err := rollbackLastMigration(conn, cfg.MigrationsDir); err != nil {
			log.Fatalf("rollback failed: %v", err)
		}
		fmt.Println("Last migration rolled back.")
	default:
		log.Fatalf("unknown mode %q, use migrate or rollback", *mode)
	}
}

func runMigrations(dbConn *sql.DB, dir string) error {
	// создаём таблицу (если ещё нет) с dirty=false по умолчанию
	createTable := `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    dirty   BOOLEAN NOT NULL DEFAULT FALSE
);`
	if _, err := dbConn.Exec(createTable); err != nil {
		return fmt.Errorf("failed to ensure schema_migrations table: %w", err)
	}
	// на случай, если столбец уже был без DEFAULT
	if _, err := dbConn.Exec(`ALTER TABLE schema_migrations ALTER COLUMN dirty SET DEFAULT FALSE`); err != nil {
		return fmt.Errorf("failed to alter default on dirty: %w", err)
	}
	if _, err := dbConn.Exec(`UPDATE schema_migrations SET dirty = FALSE WHERE dirty IS NULL`); err != nil {
		return fmt.Errorf("failed to backfill dirty flags: %w", err)
	}

	// список .up.sql
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to list migration files: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		name := filepath.Base(file)
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			return fmt.Errorf("invalid migration filename: %s", name)
		}
		version, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid version prefix %q: %w", parts[0], err)
		}

		var exists bool
		if err := dbConn.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`,
			version,
		).Scan(&exists); err != nil {
			return fmt.Errorf("failed to check migration %d: %w", version, err)
		}
		if exists {
			log.Printf("skip %d (%s)", version, name)
			continue
		}

		// выполняем файл .up.sql
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", name, err)
		}
		for _, stmt := range strings.Split(string(content), ";") {
			if s := strings.TrimSpace(stmt); s != "" {
				if _, err := dbConn.Exec(s); err != nil {
					return fmt.Errorf("error in %s: %w", name, err)
				}
			}
		}

		// отмечаем как применённую
		if _, err := dbConn.Exec(
			`INSERT INTO schema_migrations(version, dirty) VALUES($1, FALSE)`, version,
		); err != nil {
			return fmt.Errorf("failed to record migration %d: %w", version, err)
		}

		log.Printf("applied %d (%s)", version, name)
	}

	return nil
}

func rollbackLastMigration(dbConn *sql.DB, dir string) error {
	// достаём последнюю версию
	var version int64
	err := dbConn.QueryRow(`SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`).Scan(&version)
	if err == sql.ErrNoRows {
		return fmt.Errorf("no migrations to rollback")
	}
	if err != nil {
		return fmt.Errorf("failed to query last migration: %w", err)
	}

	// ищем .down.sql
	verStr := fmt.Sprintf("%04d", version)
	pattern := filepath.Join(dir, verStr+"_*.down.sql")
	matches, _ := filepath.Glob(pattern)
	if len(matches) == 0 {
		return fmt.Errorf("down-file for version %d not found", version)
	}
	downFile := matches[0]

	// выполняем down
	bytes, err := os.ReadFile(downFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", downFile, err)
	}
	for _, stmt := range strings.Split(string(bytes), ";") {
		if s := strings.TrimSpace(stmt); s != "" {
			if _, err := dbConn.Exec(s); err != nil {
				return fmt.Errorf("error in %s: %w", downFile, err)
			}
		}
	}

	// удаляем запись о миграции
	if _, err := dbConn.Exec(`DELETE FROM schema_migrations WHERE version = $1`, version); err != nil {
		return fmt.Errorf("failed to delete migration record %d: %w", version, err)
	}

	log.Printf("rolled back %d (%s)", version, filepath.Base(downFile))
	return nil
}
