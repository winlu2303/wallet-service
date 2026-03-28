# Wallet Service

Сервис для управления балансами кошельков с поддержкой конкурентных операций (1000 RPS).

## Функциональность

- Пополнение баланса кошелька (DEPOSIT)
- Списание средств с кошелька (WITHDRAW)
- Получение текущего баланса
- Docker контейнеризация

## Технологии

- **Go 1.21** - язык программирования
- **PostgreSQL 15** - база данных
- **Docker & Docker Compose** - контейнеризация
- **Gorilla Mux** - маршрутизация
- **Testify** - тестирование

## Требования

- Docker и Docker Compose
- Go 1.21+

## Быстрый старт

### 1. Клонирование репозитория

```bash
git clone <repository-url>
cd wallet-service
```

### 2. Запуск через Docker-Compose

# Запуск всех сервисов
docker-compose up -d

# Лучше подождать пока БД инициализируется
sleep 10

# Просмотр логов
docker-compose logs -f app

# Остановка сервисов
docker-compose down

### 3. Проверка работы

# Проверка статуса контейнеров
docker-compose ps

## Тестирование

# Запуск всех тестов
go test ./... -v

## Просмотр логов

# Все логи
docker-compose logs -f

## Подключение к базе данных 

# Подключение через psql
docker-compose exec postgres psql -U postgres -d walletdb

# Просмотр таблиц
\dt

# Просмотр данных
SELECT * FROM wallets;

# Просмотр активных транзакций
SELECT * FROM pg_stat_activity;

## Структура проекта

``` 
wallet-service/
├── cmd/
│   └── app/
│       └── main.go                 # Точка входа
├── internal/
│   ├── api/
│   │   └── handler.go              # HTTP обработчики
│   ├── models/
│   │   └── wallet.go               # Модели данных
│   ├── repository/
│   │   └── postgres.go             # Работа с БД
│   └── service/
│       └── wallet.go               # Бизнес-логика
│       └── wallet_test.go          # Тест б/л
├── migrations/
│   └── 001_create_wallets_table.sql # SQL миграции
├── docker-compose.yml              # Docker Compose конфиг
├── Dockerfile                      # Docker образ
├── config.env                      # Переменные окружения
├── go.mod                          # Go модуль
├── go.sum                          # Go зависимости
└── README.md                       # Документация
```
### Тестирование после запуска

# Запуск приложения 

## Полное тестирование API

# 1. Получить баланс
curl http://localhost:8081/api/v1/wallets/550e8400-e29b-41d4-a716-446655440000

# 2. Пополнить баланс
curl -X POST http://localhost:8081/api/v1/wallets \
  -H "Content-Type: application/json" \
  -d '{
    "walletId": "550e8400-e29b-41d4-a716-446655440000",
    "operationType": "DEPOSIT",
    "amount": 500
  }'

# 3. Списать средства 
curl -X POST http://localhost:8081/api/v1/wallets \
  -H "Content-Type: application/json" \
  -d '{
    "walletId": "550e8400-e29b-41d4-a716-446655440000",
    "operationType": "WITHDRAW",
    "amount": 300
  }'

# 4. Проверить новый баланс, повторив п.1.

## Тест конкурентности (5 параллельных запросов, но можно и 10+)
for i in {1..5}; do
  (for j in {1..5}; do
    curl -s -X POST http://localhost:8080/api/v1/wallet \
      -H "Content-Type: application/json" \
      -d '{"walletId":"550e8400-e29b-41d4-a716-446655440000","operationType":"DEPOSIT","amount":1}' &
  done) &
done
wait
echo "Done! Check balance:"
curl http://localhost:8080/api/v1/wallets/550e8400-e29b-41d4-a716-446655440000
