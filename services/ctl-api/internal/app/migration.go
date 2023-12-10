package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type MigrationStatus string

const (
	MigrationStatusApplied    MigrationStatus = "applied"
	MigrationStatusInProgress MigrationStatus = "in_progress"
	MigrationStatusError      MigrationStatus = "error"
)

type Migration struct {
	ID        string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	Name   string          `json:"name" gorm:"not null;default null;index:idx_migration_name,unique"`
	Status MigrationStatus `json:"status" gorm:"not null;default null"`
}

func (a *Migration) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewMigrationID()
	}

	return nil
}
