package app

import (
	"time"

	"gorm.io/gorm"
)

type InstallComponent struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	InstallID string
	Install   Install `faker:"-"`

	ComponentID string
	Component   Component `faker:"-"`

	InstallDeploys []*InstallDeploy `faker:"-"`
}
