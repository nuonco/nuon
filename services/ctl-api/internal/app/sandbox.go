package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type Sandbox struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time      `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"notnull"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string           `gorm:"unique" json:"name" gorm:"notnull"`
	Description string           `json:"description" gorm:"notnull"`
	Releases    []SandboxRelease `json:"releases" gorm:"constraint:OnDelete:CASCADE;"`
}

func (s *Sandbox) BeforeCreate(tx *gorm.DB) error {
	s.ID = domains.NewSandboxID()
	s.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	return nil
}
