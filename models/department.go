package models

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Подразделение
type Department struct {
	Id        uint         `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string       `json:"name" gorm:"column:name;not null;size:200"`
	ParentId  *uint        `json:"parent_id" gorm:"column:parent_id"`
	Parent    *Department  `json:"parent,omitempty" gorm:"foreignKey:ParentId"`
	Children  []Department `json:"children,omitempty" gorm:"foreignKey:ParentId"`
	CreatedAt time.Time    `json:"created_at" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
}

// Структура для создания/обновления отдела
type DepartmentRequest struct {
	Name     string `json:"name"`
	ParentID *uint  `json:"parent_id"`
}

// Структура для ответа API
type DepartmentResponse struct {
	Department
	Employees []Employee           `json:"employees,omitempty"`
	Children  []DepartmentResponse `json:"children,omitempty"`
}

// Имя для таблицы
func (Department) TableName() string {
	return "departments"
}

// Валидация
func (d *DepartmentRequest) Validate() error {
	// Пробелы
	d.Name = strings.TrimSpace(d.Name)

	// Проверка имени
	if d.Name == "" {
		return ErrNameEmpty
	}
	if len(d.Name) > 200 {
		return ErrFullNameTooLong
	}

	return nil
}

// Создает новый подразделения
func CreateDepartment(db *gorm.DB, req *DepartmentRequest) (*Department, error) {
	// Валидация
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Проверка родителья
	if req.ParentID != nil {
		var parent Department
		if err := db.First(&parent, *req.ParentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrParentNotFound
			}
			return nil, err
		}
	}

	// Создание подразделения
	department := &Department{
		Name:      req.Name,
		ParentId:  req.ParentID,
		CreatedAt: time.Now(),
	}

	if err := db.Create(department).Error; err != nil {
		// Проверка Имени
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			return nil, ErrNameExists
		}
		return nil, err
	}

	return department, nil
}

// Обновляет существующее подразделения
func UpdateDepartment(db *gorm.DB, id uint, req *DepartmentRequest) (*Department, error) {
	// Проверка существования
	var department Department
	if err := db.First(&department, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDepartmentNotFound
		}
		return nil, err
	}

	// Валидация имени
	if req.Name != "" {
		req.Name = strings.TrimSpace(req.Name)
		if len(req.Name) > 200 {
			return nil, ErrNameTooLong
		}
		department.Name = req.Name
	}

	// Обновление parent_id
	if req.ParentID != nil {
		// Проверка нового родителя
		var parent Department
		if err := db.First(&parent, *req.ParentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrParentNotFound
			}
			return nil, err
		}

		// Проверка родитель - потомок
		if err := checkCycle(db, id, *req.ParentID); err != nil {
			return nil, err
		}

		department.ParentId = req.ParentID
	}

	// Сохранение
	if err := db.Save(&department).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, ErrNameExists
		}
		return nil, err
	}

	return &department, nil
}

// Удалить подразделение с переводом сотрудников
func DeleteDepartment(db *gorm.DB, id uint, mode string, reassignToID *uint) error {
	// Проверка существования
	var department Department
	if err := db.First(&department, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDepartmentNotFound
		}
		return err
	}

	// 2. Обрабатываем режимы
	switch mode {
	case "cascade":
		// Каскадное удаление
		return db.Delete(&department).Error

	case "reassign":
		// С переводом
		if reassignToID == nil {
			return errors.New("reassign_to_department_id is required for reassign mode")
		}

		if *reassignToID == id {
			return ErrReassignToSame
		}

		// Проверяем целевое
		var targetDepartment Department
		if err := db.First(&targetDepartment, *reassignToID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTargetNotFound
			}
			return err
		}

		return db.Transaction(func(tx *gorm.DB) error {
			// Переводим сотрудников
			if err := tx.Model(&Employee{}).
				Where("department_id = ?", id).
				Update("department_id", *reassignToID).Error; err != nil {
				return err
			}

			// Получаем все дочерние
			var children []Department
			if err := tx.Where("parent_id = ?", id).Find(&children).Error; err != nil {
				return err
			}

			// Удаление дочерних с сотрудниками
			for _, child := range children {
				if err := tx.Delete(&child).Error; err != nil {
					return err
				}
			}

			// Удаляем подразделение
			if err := tx.Delete(&department).Error; err != nil {
				return err
			}

			return nil
		})

	default:
		return ErrInvalidMode
	}
}

// Проверяет нового parent_id
func checkCycle(db *gorm.DB, deptID, newParentID uint) error {
	if deptID == newParentID {
		return ErrSelfParent
	}

	// Получаем всех потомков
	var childIDs []uint
	if err := getChildIDs(db, deptID, &childIDs); err != nil {
		return err
	}

	// Проверка родитель-потомок
	for _, childID := range childIDs {
		if newParentID == childID {
			return ErrCycleDetected
		}
	}

	return nil
}

// Собирает ID потомков
func getChildIDs(db *gorm.DB, parentID uint, ids *[]uint) error {
	var children []Department
	if err := db.Where("parent_id = ?", parentID).Find(&children).Error; err != nil {
		return err
	}

	for _, child := range children {
		*ids = append(*ids, child.Id)
		if err := getChildIDs(db, child.Id, ids); err != nil {
			return err
		}
	}
	return nil
}

// Get - Получить подразделение
func (d *Department) GetWithTree(db *gorm.DB, id uint, depth int, includeEmployees bool) (*DepartmentResponse, error) {
	// Получаем сам отдел
	if err := db.First(d, id).Error; err != nil {
		return nil, err
	}

	response := &DepartmentResponse{
		Department: *d,
	}

	// Загружаем сотрудников
	if includeEmployees {
		var employees []Employee
		if err := db.Where("department_id = ?", id).
			Order("created_at DESC, full_name ASC").
			Find(&employees).Error; err != nil {
			return nil, err
		}
		response.Employees = employees
	}

	// Загружаем потомков
	if depth > 0 {
		var children []Department
		if err := db.Where("parent_id = ?", id).Find(&children).Error; err != nil {
			return nil, err
		}

		for _, child := range children {
			childResponse, err := child.GetWithTree(db, child.Id, depth-1, includeEmployees)
			if err != nil {
				return nil, err
			}
			response.Children = append(response.Children, *childResponse)
		}
	}

	return response, nil
}

// Валидация родства
func (d *Department) ValidateParent(db *gorm.DB) error {
	if d.ParentId == nil {
		return nil
	}

	// Самоссылку
	if *d.ParentId == d.Id {
		return ErrSelfParent
	}

	// Проверка потомка
	var childIds []uint
	if err := d.getChildIDs(db, d.Id, &childIds); err != nil {
		return err
	}

	for _, childID := range childIds {
		if *d.ParentId == childID {
			return ErrCycleDetected
		}
	}

	return nil
}

// Собираем ID всех потомков
func (d *Department) getChildIDs(db *gorm.DB, parentID uint, ids *[]uint) error {
	var children []Department
	if err := db.Where("parent_id = ?", parentID).Find(&children).Error; err != nil {
		return err
	}

	for _, child := range children {
		*ids = append(*ids, child.Id)
		if err := d.getChildIDs(db, child.Id, ids); err != nil {
			return err
		}
	}
	return nil
}
