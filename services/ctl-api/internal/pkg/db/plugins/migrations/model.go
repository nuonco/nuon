package migrations

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type MigrationStatus string

const (
	MigrationStatusApplied    MigrationStatus = "applied"
	MigrationStatusInProgress MigrationStatus = "in_progress"
	MigrationStatusError      MigrationStatus = "error"
)

type MigrationModel struct {
	ID        string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	Name   string          `json:"name" gorm:"unique"`
	Status MigrationStatus `json:"status" gorm:"not null;default null"`
}

func (a *MigrationModel) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewMigrationID()
	}

	return nil
}

func (a *MigrationModel) TableName() string {
	return "migrations"
}
