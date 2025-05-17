package config

import (
	"os"
)

func GetDBConnString() string {
	// например: "postgres://user:pass@localhost:5432/dbname?sslmode=disable"
	return os.Getenv("DATABASE_URL")
}
