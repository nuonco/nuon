package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type InstallComponent struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	InstallID   string    `json:"install_id" gorm:"index:install_component_group,unique"`
	Install     Install   `faker:"-"`
	ComponentID string    `json:"component_id" gorm:"index:install_component_group,unique"`
	Component   Component `faker:"-"`

	InstallDeploys []InstallDeploy `faker:"-" gorm:"constraint:OnDelete:CASCADE;"`
}

func (c *InstallComponent) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	return nil
}
