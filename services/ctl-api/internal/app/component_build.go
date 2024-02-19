package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type ComponentBuild struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   UserToken             `json:"created_by" gorm:"references:Subject"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	ComponentConfigConnectionID string                    `json:"component_config_connection_id" gorm:"notnull"`
	ComponentConfigConnection   ComponentConfigConnection `json:"-"`

	VCSConnectionCommitID *string              `json:"-"`
	VCSConnectionCommit   *VCSConnectionCommit `json:"vcs_connection_commit"`

	ComponentReleases []ComponentRelease `json:"releases" gorm:"constraint:OnDelete:CASCADE;"`
	InstallDeploys    []InstallDeploy    `json:"install_deploys" gorm:"constraint:OnDelete:CASCADE;"`

	Status            string  `json:"status" gorm:"notnull"`
	StatusDescription string  `json:"status_description" gorm:"notnull"`
	GitRef            *string `json:"git_ref"`

	// Read-only fields set on the object to de-nest data
	ComponentID   string `gorm:"-" json:"component_id"`
	ComponentName string `gorm:"-" json:"component_name"`
}

func (c *ComponentBuild) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewBuildID()
	if c.CreatedByID == "" {
		c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (c *ComponentBuild) AfterQuery(tx *gorm.DB) error {
	c.ComponentID = c.ComponentConfigConnection.ComponentID
	c.ComponentName = c.ComponentConfigConnection.Component.Name
	return nil
}
