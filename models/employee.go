package models

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Сотрудник
type Employee struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	DepartmentId uint       `json:"department_id" gorm:"column:department_id;not null"`
	FullName     string     `json:"full_name" gorm:"column:full_name;not null;size:200"`
	Position     string     `json:"position" gorm:"column:position;not null;size:200"`
	HiredAt      *time.Time `json:"hired_at" gorm:"column:hired_at"`
	CreatedAt    time.Time  `json:"created_at" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
}

// Структура для создания сотрудника
type EmployeeRequest struct {
	FullName string     `json:"full_name"`
	Position string     `json:"position"`
	HiredAt  *time.Time `json:"hired_at"`
}

// Структура ответа
type EmployeeResponse struct {
	Employee
	DepartmentName string `json:"department_name,omitempty"`
}

// Имя для таблицы
func (Employee) TableName() string {
	return "employees"
}

// Валидация Сотрудника
func (e *EmployeeRequest) Validate() error {
	// Пробелов
	e.FullName = strings.TrimSpace(e.FullName)
	e.Position = strings.TrimSpace(e.Position)

	// Проверка имени
	if e.FullName == "" {
		return ErrFullNameEmpty
	}
	if len(e.FullName) > 200 {
		return ErrFullNameTooLong
	}

	// Проверка должности
	if e.Position == "" {
		return ErrPositionEmpty
	}
	if len(e.Position) > 200 {
		return ErrPositionTooLong
	}

	// Проверка даты найма
	if e.HiredAt != nil && e.HiredAt.After(time.Now()) {
		return ErrHiredAtFuture
	}

	return nil
}

// CreateEmployee создает нового сотрудника в указанном отделе
func CreateEmployee(db *gorm.DB, departmentID uint, req *EmployeeRequest) (*Employee, error) {
	// Проверка отдела
	var department Department
	if err := db.First(&department, departmentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDepartmentNotFound
		}
		return nil, err
	}

	// Вызов валидации
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Создание
	employee := &Employee{
		DepartmentId: departmentID,
		FullName:     req.FullName,
		Position:     req.Position,
		HiredAt:      req.HiredAt,
		CreatedAt:    time.Now(),
	}

	if err := db.Create(employee).Error; err != nil {
		return nil, err
	}

	return employee, nil
}

// Get - всех сотрудников в отделе
func GetEmployeesByDepartment(db *gorm.DB, departmentID uint, sortBy string) ([]Employee, error) {
	var employees []Employee

	query := db.Where("department_id = ?", departmentID)

	// Сортировка
	switch sortBy {
	case "name":
		query = query.Order("full_name ASC")
	case "created":
		fallthrough
	default:
		query = query.Order("created_at DESC")
	}

	if err := query.Find(&employees).Error; err != nil {
		return nil, err
	}

	return employees, nil
}

// Перемещение сотрудника между отделами
func MoveEmployees(db *gorm.DB, fromDeptID, toDeptID uint) error {
	return db.Model(&Employee{}).
		Where("department_id = ?", fromDeptID).
		Update("department_id", toDeptID).Error
}
