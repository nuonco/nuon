package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type Component struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26;" json:"id"`
	CreatedByID string         `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time      `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"notnull"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`

	Name string `json:"name" gorm:"notnull;index:idx_app_component_name,unique"`

	AppID string `json:"app_id" gorm:"notnull;index:idx_app_component_name,unique"`
	App   App    `faker:"-" json:"-"`

	ConfigVersions   int                         `gorm:"-" json:"config_versions"`
	ComponentConfigs []ComponentConfigConnection `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
}

func (c *Component) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
