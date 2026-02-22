package models

import (
	"time"

	"gorm.io/gorm"
)

// Employee — сотрудник
type Employee struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	DepartmentId uint       `json:"department_id" gorm:"column:department_id"`
	Department   Department `gorm:"foreignKey:DepartmentId"`
	FullName     string     `json:"full_name" gorm:"column:full_name"`
	Position     string     `json:"position" gorm:"column:position"`
	HiredAt      *time.Time `json:"hired_at" gorm:"column:hired_at"`
	CreatedAt    time.Time  `json:"created_at" gorm:"column:created_at"`
}

func MigrateEmployee(db *gorm.DB) error {
	return db.AutoMigrate(&Employee{})
}
