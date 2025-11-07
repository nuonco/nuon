package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type ConnectedGithubVCSConfig struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	// parent component
	ComponentConfigID   string `json:"component_config_id,omitzero" gorm:"notnull" temporaljson:"component_config_id,omitzero,omitempty"`
	ComponentConfigType string `json:"component_config_type,omitzero" gorm:"notnull" temporaljson:"component_config_type,omitzero,omitempty"`

	Repo      string `json:"repo,omitzero" gorm:"notnull" temporaljson:"repo,omitzero,omitempty"`
	RepoName  string `json:"repo_name,omitzero" gorm:"notnull" temporaljson:"repo_name,omitzero,omitempty"`
	RepoOwner string `json:"repo_owner,omitzero" gorm:"notnull" temporaljson:"repo_owner,omitzero,omitempty"`
	Directory string `json:"directory,omitzero" gorm:"notnull" temporaljson:"directory,omitzero,omitempty"`
	Branch    string `json:"branch,omitzero" gorm:"notnull" temporaljson:"branch,omitzero,omitempty"`

	VCSConnectionID string        `json:"vcs_connection_id,omitzero" gorm:"notnull" temporaljson:"vcs_connection_id,omitzero,omitempty"`
	VCSConnection   VCSConnection `json:"vcs_connection,omitzero,omitempty" temporaljson:"vcs_connection,omitzero,omitempty"`
}

func (c *ConnectedGithubVCSConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &ConnectedGithubVCSConfig{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (c *ConnectedGithubVCSConfig) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewVCSID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
