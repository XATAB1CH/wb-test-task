# Базовый образ с Go
FROM golang:1.22-alpine AS builder

# Переменные окружения
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Рабочая директория
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Сборка бинарника
RUN go build -o wb-service ./cmd/main.go

# Минимальный образ для запуска
FROM alpine:latest

WORKDIR /root/

# Копируем бинарник из стадии сборки
COPY --from=builder /app/wb-service .

# Копируем .env (можно убрать, если хочешь передавать всё через переменные окружения)
COPY .env .env

# Порт HTTP сервера
EXPOSE 8081

# Запуск приложения
CMD ["./wb-service"]