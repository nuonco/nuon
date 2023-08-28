package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type InstallComponent struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time      `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"notnull"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`

	InstallID   string    `json:"install_id" gorm:"index:install_component_group,unique;notnull"`
	Install     Install   `faker:"-"`
	ComponentID string    `json:"component_id" gorm:"index:install_component_group,unique;notnull"`
	Component   Component `faker:"-"`

	InstallDeploys []InstallDeploy `faker:"-" gorm:"constraint:OnDelete:CASCADE;"`
}

func (c *InstallComponent) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
