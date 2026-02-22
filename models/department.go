package models

import (
	"time"

	"gorm.io/gorm"
)

// Department - подразделение
type Department struct {
	Id        uint         `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string       `json:"name" gorm:"column:name;not null"`
	ParentId  *uint        `json:"parent_id" gorm:"column:parent_id"`
	Parent    *Department  `gorm:"foreignKey:ParentId"`
	Children  []Department `gorm:"foreignKey:ParentId"`
	CreatedAt time.Time    `json:"created_at" gorm:"column:created_at"`
}

func MigrateDepartment(db *gorm.DB) error {
	return db.AutoMigrate(&Department{})
}
