package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type DockerBuildComponentConfig struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`

	// value
	ComponentConfigConnectionID string                    `json:"component_config_connection_id" gorm:"notnull"`
	ComponentConfigConnection   ComponentConfigConnection `json:"-"`

	Dockerfile string                   `json:"dockerfile" gorm:"default:Dockerfile;notnull"`
	Target     string                   `json:"target"`
	BuildArgs  pgtype.FlatArray[string] `gorm:"type:text[]" json:"build_args" swaggertype:"array,string"`
	EnvVars    pgtype.Hstore            `json:"env_vars" gorm:"type:hstore" swaggertype:"object,string"`

	PublicGitVCSConfig       *PublicGitVCSConfig       `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"public_git_vcs_config,omitempty"`
	ConnectedGithubVCSConfig *ConnectedGithubVCSConfig `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"connected_github_vcs_config,omitempty"`
}

func (c *DockerBuildComponentConfig) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewConfigID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
