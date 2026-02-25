package main

import (
	"net/http"
	"os"

	"github.com/kroulersama/goProject/internal/handler"

	"github.com/kroulersama/goProject/pkg/logger"
	"github.com/kroulersama/goProject/storage"
	_ "github.com/lib/pq"
)

func main() {
	// Инициализация логгера
	log := logger.New()

	// Логируем запуск
	log.Info("Starting server...")

	//Инициализация базы с goose миграциями
	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASSWORD"),
		User:     os.Getenv("DB_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   os.Getenv("DB_NAME"),
	}

	dsn := config.GetDSN()

	log.Info("Waiting for database...")
	handler.WaitForDB(dsn)
	log.Info("Database connected")

	if err := handler.RunMigrations(dsn); err != nil {
		log.Fatal("Failed to run migrations", err)
	}
	log.Info("Migrations applied successfully")

	//Инициализация GORM
	db, err := storage.NewConnection(config)
	if err != nil {
		log.Fatal("Could not load the database", err)
	}
	log.Info("GORM initialized")
	repo := &handler.Repository{
		DB:  db,
		Log: log,
	}

	//Инициализация Путей
	mux := http.NewServeMux()
	mux.HandleFunc("POST /departments", log.Middleware(repo.CreateDepartment))
	mux.HandleFunc("POST /departments/{id}/employees", log.Middleware(repo.CreateEmployeeInDepartment))
	mux.HandleFunc("GET /departments/{id}", log.Middleware(repo.GetDepartment))
	mux.HandleFunc("PATCH /departments/{id}", log.Middleware(repo.MoveDepartment))
	mux.HandleFunc("DELETE /departments/{id}", log.Middleware(repo.DeleteDepartment))

	log.Info("Server started on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("Server failed", err)
	}
}
