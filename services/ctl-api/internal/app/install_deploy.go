package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type InstallDeploy struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	BuildID string         `json:"build_id"`
	Build   ComponentBuild `faker:"-" json:"build,omitempty"`

	InstallComponentID string           `json:"install_component_id"`
	InstallComponent   InstallComponent `faker:"-" json:"-"`

	Status            string `json:"status"`
	StatusDescription string `json:"status_description"`
}

func (c *InstallDeploy) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewDeployID()
	return nil
}
