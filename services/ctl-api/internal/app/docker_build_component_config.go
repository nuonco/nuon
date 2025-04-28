package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type DockerBuildComponentConfig struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	// value
	ComponentConfigConnectionID string                    `json:"component_config_connection_id" gorm:"notnull" temporaljson:"component_config_connection_id,omitzero,omitempty"`
	ComponentConfigConnection   ComponentConfigConnection `json:"-" temporaljson:"component_config_connection,omitzero,omitempty"`

	Dockerfile string         `json:"dockerfile" gorm:"default:Dockerfile;notnull" temporaljson:"dockerfile,omitzero,omitempty"`
	Target     string         `json:"target" temporaljson:"target,omitzero,omitempty"`
	BuildArgs  pq.StringArray `gorm:"type:text[]" json:"build_args" swaggertype:"array,string" temporaljson:"build_args,omitzero,omitempty"`
	EnvVars    pgtype.Hstore  `json:"env_vars" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"env_vars,omitzero,omitempty"`

	PublicGitVCSConfig       *PublicGitVCSConfig       `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"public_git_vcs_config,omitempty" temporaljson:"public_git_vcs_config,omitzero,omitempty"`
	ConnectedGithubVCSConfig *ConnectedGithubVCSConfig `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"connected_github_vcs_config,omitempty" temporaljson:"connected_github_vcs_config,omitzero,omitempty"`
}

func (c *DockerBuildComponentConfig) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewConfigID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
