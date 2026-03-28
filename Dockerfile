#Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app

#Устанавливаем зависимости
COPY go.mod go.sum ./
RUN go mod download

#Копируем исходный код
COPY . .

#Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o wallet-service ./cmd/wallet-service

#Используем минимальный образ для финального контейнера
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/wallet-service .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./wallet-service"]