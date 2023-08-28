package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type ComponentBuild struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time      `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"notnull"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`

	ComponentConfigConnectionID string                    `json:"component_config_connection_id" gorm:"notnull"`
	ComponentConfigConnection   ComponentConfigConnection `json:"-"`

	VCSConnectionCommitID *string              `json:"-"`
	VCSConnectionCommit   *VCSConnectionCommit `json:"vcs_connection_commit"`

	ComponentReleases []ComponentRelease `json:"releases" gorm:"constraint:OnDelete:CASCADE;"`

	Status            string  `json:"status" gorm:"notnull"`
	StatusDescription string  `json:"status_description" gorm:"notnull"`
	GitRef            *string `json:"git_ref"`
}

func (c *ComponentBuild) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
