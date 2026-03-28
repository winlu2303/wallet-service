package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"wallet-service/internal/api"
	"wallet-service/internal/repository"
	"wallet-service/internal/service"
)

func main() {
	// Загружаем переменные окружения из .env файла
	if err := godotenv.Load("config.env"); err != nil {
		log.Printf("Warning: config.env file not found")

	}
	//Подключение к базе данных PostgreSQL
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/wallet_db?sslmode=disable"
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	//Проверяем соединение с базой данных
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	//Настраиваем репозиторий, сервис и обработчики
	repo := repository.NewPostgresRepository(db)
	walletService := service.NewWalletService(repo)
	handler := api.NewWalletHandler(walletService)

	//Настраиваем маршруты и запускаем HTTP сервер (роутер)
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/wallets", handler.ProcessOperation).Methods("POST")
	router.HandleFunc("/api/v1/wallets/{id}", handler.GetBalance).Methods("GET")

	//Запускаем сервер
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
