package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/kroulersama/goProject/models"
	"github.com/kroulersama/goProject/pkg/logger"
	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
)

// Применяем миграции
func RunMigrations(dsn string) error {
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
func WaitForDB(dsn string) {
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
	DB  *gorm.DB
	Log *logger.Logger
}

// Тип для запроса подразделения
type CreateDepartmentRequest = models.DepartmentRequest

// CreateDepartment создание предприятия
func (r *Repository) CreateDepartment(w http.ResponseWriter, req *http.Request) {
	r.Log.Info("Creating department", "method", req.Method)

	// Проверка метода
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Декодирование JSON
	var deptReq models.DepartmentRequest
	if err := json.NewDecoder(req.Body).Decode(&deptReq); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "request failed",
			"error":   err.Error(),
		})
		return
	}

	// Обработка в модуле
	department, err := models.CreateDepartment(r.DB, &deptReq)
	if err != nil {
		r.Log.Error("Failed to create department", err, "name", deptReq.Name)

		switch {
		case errors.Is(err, models.ErrNameEmpty) ||
			errors.Is(err, models.ErrNameTooLong) ||
			errors.Is(err, models.ErrParentNotFound):
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})

		case errors.Is(err, models.ErrNameExists):
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

	r.Log.Info("Department created", "id", department.Id, "name", department.Name)

	// Успешный ответ
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "department created successfully",
		"data":    department,
	})
}

// Тип для запроса сотрудника
type CreateEmployeeRequest = models.EmployeeRequest

// CreateEmployeeInDepartment создание сотрудника
func (r *Repository) CreateEmployeeInDepartment(w http.ResponseWriter, req *http.Request) {
	r.Log.Info("Creating employee", "method", req.Method)

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
		r.Log.Error("Failed to create employee", err, "name", empReq.FullName)

		switch {
		case errors.Is(err, models.ErrDepartmentNotFound):
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})

		case errors.Is(err, models.ErrFullNameEmpty),
			errors.Is(err, models.ErrFullNameTooLong),
			errors.Is(err, models.ErrPositionEmpty),
			errors.Is(err, models.ErrPositionTooLong),
			errors.Is(err, models.ErrHiredAtFuture):
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

	r.Log.Info("employee created", "id", employee.ID, "full_name", employee.FullName)

	// Ответ
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "employee created successfully",
		"data":    employee,
	})
}

// GetDepartment вызов информации о подразделении
func (r *Repository) GetDepartment(w http.ResponseWriter, req *http.Request) {
	r.Log.Info("Getting department", "method", req.Method)

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
		r.Log.Error("Failed get department", err, "name", dept.Name)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, `{"message": "department not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"message": "database error"}`, http.StatusInternalServerError)
		return
	}

	r.Log.Info("Department", "id", dept.Id, "name", dept.Name)

	// Ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// MoveDepartment перемещение подразделения с изменением родителя
func (r *Repository) MoveDepartment(w http.ResponseWriter, req *http.Request) {
	r.Log.Info("moving department", "method", req.Method)

	// Проверка метода
	if req.Method != http.MethodPatch {
		http.Error(w, `{"message": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Вычленение Id
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

	// Обработка запроса
	var deptReq models.DepartmentRequest
	if err := json.NewDecoder(req.Body).Decode(&deptReq); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "invalid request format",
			"error":   err.Error(),
		})
		return
	}

	// Логика в модуле
	updatedDepartment, err := models.UpdateDepartment(r.DB, uint(departmentID), &deptReq)
	if err != nil {
		r.Log.Error("Failed move department", err, "name", deptReq.Name)
		switch {
		case errors.Is(err, models.ErrDepartmentNotFound):
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})

		case errors.Is(err, models.ErrNameTooLong),
			errors.Is(err, models.ErrNameEmpty),
			errors.Is(err, models.ErrParentNotFound):
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})

		case errors.Is(err, models.ErrSelfParent),
			errors.Is(err, models.ErrCycleDetected),
			errors.Is(err, models.ErrNameExists):
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})

		default:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "could not update department",
				"error":   err.Error(),
			})
		}
		return
	}

	r.Log.Info("Department move", "parent_id", deptReq.ParentID, "name", deptReq.Name)

	// Успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "department updated successfully",
		"data":    updatedDepartment,
	})
}

func (r *Repository) DeleteDepartment(w http.ResponseWriter, req *http.Request) {
	r.Log.Info("del department", "method", req.Method)

	// Проверка метода
	if req.Method != http.MethodDelete {
		http.Error(w, `{"message": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Получаем Id
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

	// Параметры
	mode := req.URL.Query().Get("mode")
	if mode == "" {
		http.Error(w, `{"message": "mode parameter is required (cascade or reassign)"}`, http.StatusBadRequest)
		return
	}

	var reassignToID *uint
	if mode == "reassign" {
		reassignStr := req.URL.Query().Get("reassign_to_department_id")
		if reassignStr == "" {
			http.Error(w, `{"message": "reassign_to_department_id is required for reassign mode"}`, http.StatusBadRequest)
			return
		}

		reassignID, err := strconv.ParseUint(reassignStr, 10, 32)
		if err != nil {
			http.Error(w, `{"message": "invalid reassign_to_department_id"}`, http.StatusBadRequest)
			return
		}
		reassignIDUint := uint(reassignID)
		reassignToID = &reassignIDUint
	}

	// Логика в  модели
	err = models.DeleteDepartment(r.DB, uint(departmentID), mode, reassignToID)
	if err != nil {
		r.Log.Error("Failed del department", err, "id", departmentID, "mode", mode)
		switch {
		case errors.Is(err, models.ErrDepartmentNotFound):
			http.Error(w, `{"message": "department not found"}`, http.StatusNotFound)

		case errors.Is(err, models.ErrTargetNotFound):
			http.Error(w, `{"message": "target department not found"}`, http.StatusNotFound)

		case errors.Is(err, models.ErrReassignToSame),
			errors.Is(err, models.ErrInvalidMode):
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})

		default:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "could not delete department",
				"error":   err.Error(),
			})
		}
		return
	}
	r.Log.Info("Department del", "departmentID", departmentID, "mode", mode)

	// Успешное удаление
	w.WriteHeader(http.StatusNoContent)
}
