package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kroulersama/goProject/internal/handler"
	"github.com/kroulersama/goProject/models"
	"github.com/kroulersama/goProject/pkg/logger"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB создает тестовую БД в памяти
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Создаем таблицы
	err = db.AutoMigrate(&models.Department{}, &models.Employee{})
	require.NoError(t, err)

	return db
}

// setupTestHandler создает тестовый репозиторий и хендлер
func setupTestHandler(t *testing.T) (*handler.Repository, *httptest.Server) {
	db := setupTestDB(t)
	log := logger.New()

	repo := &handler.Repository{
		DB:  db,
		Log: log,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /departments", repo.CreateDepartment)
	mux.HandleFunc("POST /departments/{id}/employees", repo.CreateEmployeeInDepartment)
	mux.HandleFunc("GET /departments/{id}", repo.GetDepartment)
	mux.HandleFunc("PATCH /departments/{id}", repo.MoveDepartment)
	mux.HandleFunc("DELETE /departments/{id}", repo.DeleteDepartment)

	server := httptest.NewServer(mux)
	return repo, server
}
