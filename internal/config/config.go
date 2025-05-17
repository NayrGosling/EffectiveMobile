package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	DatabaseURL   string
	MigrationsDir string
	LogLevel      log.Level
	Port          int
}

// Load загружает конфигурацию из переменных окружения.
// Если переменные окружения не установлены, используются значения по умолчанию.
func Load() *Config {
	// Загружаем переменные окружения из файла .env
	_ = godotenv.Load()

	// Получаем уровень логирования из переменной окружения LOG_LEVEL
	// Если переменная не установлена или содержит некорректное значение, используем уровень Info
	lvl, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		lvl = log.InfoLevel
	}

	// Получаем номер порта из переменной окружения PORT
	// Если переменная не установлена или содержит некорректное значение, используем порт 8080
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = 8080
	}

	// Возвращаем структуру Config с загруженными значениями
	return &Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		MigrationsDir: os.Getenv("MIGRATIONS_DIR"),
		LogLevel:      lvl,
		Port:          port,
	}
}
