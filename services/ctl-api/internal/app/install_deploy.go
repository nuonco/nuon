package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type InstallDeployType string

const (
	InstallDeployTypeRelease  InstallDeployType = "release"
	InstallDeployTypeInstall  InstallDeployType = "install"
	InstallDeployTypeTeardown InstallDeployType = "teardown"
)

type InstallDeploy struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`

	ComponentBuildID string         `json:"build_id" gorm:"notnull"`
	ComponentBuild   ComponentBuild `faker:"-" json:"-"`

	InstallComponentID string           `json:"install_component_id" gorm:"notnull"`
	InstallComponent   InstallComponent `faker:"-" json:"-"`

	ComponentReleaseStepID *string               `json:"release_id"`
	ComponentReleaseStep   *ComponentReleaseStep `json:"-"`

	Status            string            `json:"status" gorm:"notnull"`
	StatusDescription string            `json:"status_description" gorm:"notnull"`
	Type              InstallDeployType `json:"install_deploy_type"`

	// Fields that are de-nested at read time
	InstallID     string `json:"install_id" gorm:"-"`
	ComponentID   string `json:"component_id" gorm:"-"`
	ComponentName string `json:"component_name" gorm:"-"`
}

func (c *InstallDeploy) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewDeployID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
