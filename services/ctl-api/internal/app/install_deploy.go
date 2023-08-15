package app

import (
	"time"

	"gorm.io/gorm"
)

type InstallDeploy struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	BuildID string
	Build   ComponentBuild `faker:"-"`

	InstallComponentID string
	InstallComponent   InstallComponent `faker:"-"`
}
