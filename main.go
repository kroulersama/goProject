package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/kroulersama/goProject/models"
	"github.com/kroulersama/goProject/storage"
	"gorm.io/gorm"
)

func main() {
	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASSWORD"),
		User:     os.Getenv("DB_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   os.Getenv("DB_NAME"),
	}

	db, err := storage.NewConnection(config)
	if err != nil {
		log.Fatal("could not load the database")
	}

	err = models.MigrateDepartment(db)
	if err != nil {
		log.Fatal("could not migrate db")
	}

	r := Reposytory{
		DB: db,
	}

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

type Reposytory struct {
	DB *gorm.DB
}

// Тип для запроса подразделения
type CreateDepartmentRequest struct {
	Name     string `json:"name"`
	ParentID *uint  `json:"parent_id"`
}

func (r *Reposytory) CreateDepartment(w http.ResponseWriter, req *http.Request) {

	//Обработка заявки
	if req.Method != http.MethodPost {
		http.Error(w, "Mettod not allowed", http.StatusMethodNotAllowed)
		return
	}

	var deptReq CreateDepartmentRequest

	if err := json.NewDecoder(req.Body).Decode(&deptReq); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "request failed",
			"error":   err.Error(),
		})
	}

	//Валидация
	if deptReq.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "department name cannot be empty",
		})
		return
	}

	if len(deptReq.Name) > 200 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "department name too long (max 200)",
		})
		return
	}

	//Присваивание
	department := models.Department{
		Name:      deptReq.Name,
		ParentId:  deptReq.ParentID,
		CreatedAt: time.Now(),
	}

	if err := r.DB.Create(&department).Error; err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "could not create department",
			"error":   err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "department created successfully",
		"data":    department,
	})
	return
}

// Тип для запроса сотрудника
type CreateEmployeeRequest struct {
	FullName string     `json:"full_name"`
	Position string     `json:"position"`
	HiredAt  *time.Time `json:"hired_at"`
}

func (r *Reposytory) CreateEmployeeInDepartment(w http.ResponseWriter, req *http.Request) {

	//Выявление ид подразделения
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

	var department models.Department
	if err := r.DB.First(&department, uint(departmentID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, `{"message": "department not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"message": "database error"}`, http.StatusInternalServerError)
		return
	}

	//Обработка заявки
	var empReq CreateEmployeeRequest
	if err := json.NewDecoder(req.Body).Decode(&empReq); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "invalid request format",
			"error":   err.Error(),
		})
		return
	}

	//Валидация
	empReq.FullName = strings.TrimSpace(empReq.FullName)
	if empReq.FullName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "full name cannot be empty",
		})
		return
	}
	if len(empReq.FullName) > 200 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "full name too long (max 200 characters)",
		})
		return
	}

	empReq.Position = strings.TrimSpace(empReq.Position)
	if empReq.Position == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "position cannot be empty",
		})
		return
	}
	if len(empReq.Position) > 200 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "position too long (max 200 characters)",
		})
		return
	}

	//Присваивание
	employee := models.Employee{
		DepartmentId: uint(departmentID),
		FullName:     empReq.FullName,
		Position:     empReq.Position,
		HiredAt:      empReq.HiredAt,
		CreatedAt:    time.Now(),
	}

	if err := r.DB.Create(&employee).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "could not create employee",
			"error":   err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "employee created successfully",
		"data":    employee,
	})
}

func (r *Reposytory) GetDepartment(context fiber.Ctx) error {
	departmentmodels := &[]models.Department{}

	err := r.DB.Find(departmentmodels).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "Could not get department"})
		return err

	}

	context.Status(http.StatusOK).JSON(fiber.Map{
		"message": "department fetched successfully",
		"data":    departmentmodels,
	})
	return nil
}

func (r *Reposytory) MoveDepartment(context fiber.Ctx) error {
	//todo
	return nil
}

func (r *Reposytory) DeleteDepartment(context fiber.Ctx) error {
	departmentmodels := &[]models.Department{}

	id := context.Params("id")
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty"})
		return nil
	}

	err := r.DB.Delete(departmentmodels, id)
	if err.Error != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"massage": "could not delete department"})
		return err.Error
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "department delete successfully"})
	return nil
}
