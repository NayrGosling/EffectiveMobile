# Миграции
```
$env:MIGRATIONS_DIR = "$PWD\migrations"
$env:DATABASE_URL = "postgres://postgres:1111@localhost:5432/personsdb?sslmode=disable"
go run cmd/migrate/main.go
```
# Сборка в докере
```
docker-compose up --build -d
```