package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type ComponentBuild struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	ComponentConfigConnectionID string                    `json:"component_config_connection_id"`
	ComponentConfigConnection   ComponentConfigConnection `json:"-"`

	VCSConnectionCommitID *string              `json:"-"`
	VCSConnectionCommit   *VCSConnectionCommit `json:"vcs_connection_commit"`

	Status string  `json:"status"`
	GitRef *string `json:"git_ref"`
}

func (c *ComponentBuild) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	return nil
}
