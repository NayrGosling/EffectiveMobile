# --- builder stage ---
FROM golang:1.23-alpine AS builder

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o person-service cmd/server/main.go

# --- final stage ---
FROM alpine:latest

# Для работы TLS
RUN apk add --no-cache ca-certificates

WORKDIR /root/

COPY --from=builder /app/person-service .

EXPOSE 8080

CMD ["./person-service"]
