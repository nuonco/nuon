package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type ConnectedGithubVCSConfig struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// parent component
	ComponentConfigID   string `json:"component_config_id"`
	ComponentConfigType string `json:"component_config_type"`

	Repo      string `json:"repo"`
	Directory string `json:"directory"`
	Branch    string `json:"branch"`
	GitRef    string `json:"git_ref"`

	VCSConnection   VCSConnection `json:"-"`
	VCSConnectionID string        `json:"vcs_connection_id"`
}

func (c *ConnectedGithubVCSConfig) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewVCSID()
	return nil
}
