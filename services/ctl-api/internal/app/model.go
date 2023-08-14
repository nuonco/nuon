package app

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID          string         `gorm:"primary_key;" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
