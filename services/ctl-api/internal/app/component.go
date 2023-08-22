package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type Component struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time      `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"notnull"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Name string `json:"name" gorm:"notnull"`

	AppID string `json:"app_id" gorm:"notnull"`
	App   App    `faker:"-" json:"-"`

	ConfigVersions   int                         `gorm:"-" json:"config_versions"`
	ComponentConfigs []ComponentConfigConnection `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
}

func (c *Component) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	return nil
}
