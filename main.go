package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/kroulersama/goProject/models"
	"github.com/kroulersama/goProject/storage"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
)

func main() {
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

	waitForDB(dsn)

	if err := runMigrations(dsn); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	//Инициализация GORM
	db, err := storage.NewConnection(config)
	if err != nil {
		log.Fatal("could not load the database")
	}

	r := Repository{
		DB: db,
	}

	//Инициализация Путей
	mux := http.NewServeMux()
	mux.HandleFunc("POST /departments", r.CreateDepartment)
	mux.HandleFunc("POST /departments/{id}/employee", r.CreateEmployeeInDepartment)
	mux.HandleFunc("GET /departments/{id}", r.GetDepartment)
	mux.HandleFunc("PATCH /departments/{id}", r.MoveDepartment)
	mux.HandleFunc("DELETE /departments/{id}", r.DeleteDepartment)

	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

// Применяем миграции
func runMigrations(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	log.Println("Migrations applied successfully")
	return nil
}

// Ожидание готовности базы
func waitForDB(dsn string) {
	log.Println("Waiting for database...")

	for i := 0; i < 30; i++ {
		db, err := sql.Open("postgres", dsn)
		if err == nil {
			if err = db.Ping(); err == nil {
				db.Close()
				log.Println("Database is ready!")
				return
			}
			db.Close()
		}
		time.Sleep(2 * time.Second)
		log.Printf("Retrying... (%d/30)", i+1)
	}

	log.Fatal("Database not ready after 30 attempts")
}

type Repository struct {
	DB *gorm.DB
}

// Тип для запроса подразделения
type CreateDepartmentRequest struct {
	Name     string `json:"name"`
	ParentID *uint  `json:"parent_id"`
}

func (r *Repository) CreateDepartment(w http.ResponseWriter, req *http.Request) {
	// 1. Проверка метода
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Декодирование JSON
	var deptReq models.DepartmentRequest
	if err := json.NewDecoder(req.Body).Decode(&deptReq); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "request failed",
			"error":   err.Error(),
		})
		return
	}

	// 3. ВСЯ БИЗНЕС-ЛОГИКА В МОДЕЛИ
	department, err := models.CreateDepartment(r.DB, &deptReq)
	if err != nil {
		switch {
		case err.Error() == "department name cannot be empty" ||
			err.Error() == "department name too long (max 200)":
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})

		case err.Error() == "parent department not found":
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})

		case err.Error() == "department with this name already exists in this parent":
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})

		default:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "could not create department",
				"error":   err.Error(),
			})
		}
		return
	}

	// 4. Успешный ответ
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "department created successfully",
		"data":    department,
	})
}

// Тип для запроса сотрудника
type CreateEmployeeRequest = models.EmployeeRequest

func (r *Repository) CreateEmployeeInDepartment(w http.ResponseWriter, req *http.Request) {
	// Проверка метода
	if req.Method != http.MethodPost {
		http.Error(w, `{"message": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Получение id
	departmentIDStr := req.PathValue("id")
	if departmentIDStr == "" {
		http.Error(w, `{"message": "department id is required"}`, http.StatusBadRequest)
		return
	}

	departmentID, err := strconv.ParseUint(departmentIDStr, 10, 32)
	if err != nil {
		http.Error(w, `{"message": "invalid department id"}`, http.StatusBadRequest)
		return
	}

	// обработка запроса
	var empReq models.EmployeeRequest
	if err := json.NewDecoder(req.Body).Decode(&empReq); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "invalid request format",
			"error":   err.Error(),
		})
		return
	}

	// Вычисления из модуля
	employee, err := models.CreateEmployee(r.DB, uint(departmentID), &empReq)
	if err != nil {
		switch {
		case err.Error() == "department not found":
			http.Error(w, `{"message": "department not found"}`, http.StatusNotFound)
		case err.Error() == "full name cannot be empty" ||
			err.Error() == "full name too long (max 200 characters)" ||
			err.Error() == "position cannot be empty" ||
			err.Error() == "position too long (max 200 characters)" ||
			err.Error() == "hired_at cannot be in the future":
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "could not create employee",
				"error":   err.Error(),
			})
		}
		return
	}

	// Ответ
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "employee created successfully",
		"data":    employee,
	})
}

func (r *Repository) GetDepartment(w http.ResponseWriter, req *http.Request) {
	// Проверка метода
	if req.Method != http.MethodGet {
		http.Error(w, `{"message": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Получение id
	idStr := req.PathValue("id")
	if idStr == "" {
		http.Error(w, `{"message": "department id is required"}`, http.StatusBadRequest)
		return
	}

	departmentID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, `{"message": "invalid department id"}`, http.StatusBadRequest)
		return
	}

	// Обработка заявки
	depth := 1
	depthStr := req.URL.Query().Get("depth")
	if depthStr != "" {
		if d, err := strconv.Atoi(depthStr); err == nil && d >= 1 && d <= 5 {
			depth = d
		} else {
			http.Error(w, `{"message": "depth must be between 1 and 5"}`, http.StatusBadRequest)
			return
		}
	}

	includeEmployees := true
	includeStr := req.URL.Query().Get("include_employees")
	if includeStr != "" {
		if b, err := strconv.ParseBool(includeStr); err == nil {
			includeEmployees = b
		} else {
			http.Error(w, `{"message": "include_employees must be true or false"}`, http.StatusBadRequest)
			return
		}
	}

	// Вычисления из модуля
	var dept models.Department
	response, err := dept.GetWithTree(r.DB, uint(departmentID), depth, includeEmployees)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, `{"message": "department not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"message": "database error"}`, http.StatusInternalServerError)
		return
	}

	// Ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (r *Repository) MoveDepartment() {
	//todo
	return
}

func (r *Repository) DeleteDepartment() {
	// todo
	return
}
