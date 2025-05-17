# Команды для работы миграции в bash

## Применить все новые миграции
```
export DATABASE_URL="postgres://postgres:1111@localhost:5432/personsdb?sslmode=disable"
export MIGRATIONS_DIR="./migrations"
go run cmd/migrate/main.go -mode=migrate
```

## Откатить последнюю миграцию
```
export DATABASE_URL="postgres://postgres:1111@localhost:5432/personsdb?sslmode=disable"
export MIGRATIONS_DIR="./migrations"
go run cmd/migrate/main.go -mode=rollback
```

---

# Сборка в докере
```
docker-compose up --build -d
```