package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type HelmComponentConfig struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// parent reference
	ComponentConfigConnectionID string `json:"component_config_connection_id"`

	PublicGitVCSConfig       *PublicGitVCSConfig       `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"public_git_vcs_config,omitempty"`
	ConnectedGithubVCSConfig *ConnectedGithubVCSConfig `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"connected_github_vcs_config,omitempty"`

	// Helm specific configurations
	ChartName string        `json:"chart_name"`
	Values    pgtype.Hstore `json:"values" gorm:"type:hstore" swaggertype:"object,string"`
}

func (c *HelmComponentConfig) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	return nil
}
