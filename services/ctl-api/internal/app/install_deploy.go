package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type InstallDeploy struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time      `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"notnull"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	BuildID string         `json:"build_id" gorm:"notnull"`
	Build   ComponentBuild `faker:"-" json:"build,omitempty"`

	InstallComponentID string           `json:"install_component_id" gorm:"notnull"`
	InstallComponent   InstallComponent `faker:"-" json:"-"`

	ComponentReleaseStepID *string               `json:"release_id"`
	ComponentReleaseStep   *ComponentReleaseStep `json:"-"`

	Status            string `json:"status" gorm:"notnull"`
	StatusDescription string `json:"status_description" gorm:"notnull"`
}

func (c *InstallDeploy) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewDeployID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	return nil
}
