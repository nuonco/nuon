package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type DockerBuildComponentConfig struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// value
	ComponentConfigConnectionID string `json:"component_config_connection_id"`

	// VCS Config
	PublicGitVCSConfig       *PublicGitVCSConfig       `gorm:"polymorphic:ComponentConfig" json:"public_git_vcs_config,omitempty"`
	ConnectedGithubVCSConfig *ConnectedGithubVCSConfig `gorm:"polymorphic:ComponentConfig" json:"connected_github_vcs_config,omitempty"`

	SyncOnly          bool               `json:"sync_only,omitempty"`
	BasicDeployConfig *BasicDeployConfig `gorm:"polymorphic:ComponentConfig" json:"basic_deploy_config,omitempty"`

	Dockerfile string         `json:"dockerfile" gorm:"default:Dockerfile"`
	Target     string         `json:"target"`
	BuildArgs  pq.StringArray `gorm:"type:text[]" json:"build_args" swaggertype:"array,string"`
	EnvVars    pgtype.Hstore  `json:"env_vars" gorm:"type:hstore" swaggertype:"object,string"`
}

func (c *DockerBuildComponentConfig) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	return nil
}
