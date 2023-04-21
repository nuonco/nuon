package models

import (
	"time"

	"gorm.io/gorm"
)

type ModelV2 struct {
	ID        string `gorm:"primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type IDerV2 interface {
	GetID() string
}

func (m *ModelV2) GetID() string {
	return m.ID
}
