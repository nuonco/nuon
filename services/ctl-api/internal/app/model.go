package app

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID        string `gorm:"primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
