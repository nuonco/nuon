package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/config/refs"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type AppSandboxConfig struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppID       string `json:"app_id,omitzero" gorm:"not null;default null" temporaljson:"app_id,omitzero,omitempty"`
	AppConfigID string `json:"app_config_id,omitzero" temporaljson:"app_config_id,omitzero,omitempty"`

	// NOTE(jm): you can use one of a few different methods of creating an app sandbox, either a built in one, that
	// Nuon manages, or one of the public git vcs configs.

	// Either a public git repo or private repo using a connected repo source can be used. For now, these fields are
	// not being respected down stream, but will in the future.

	PublicGitVCSConfig       *PublicGitVCSConfig       `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"public_git_vcs_config,omitzero,omitempty" temporaljson:"public_git_vcs_config,omitzero,omitempty"`
	ConnectedGithubVCSConfig *ConnectedGithubVCSConfig `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"connected_github_vcs_config,omitzero,omitempty" temporaljson:"connected_github_vcs_config,omitzero,omitempty"`
	VCSConnectionType        VCSConnectionType         `json:"-" gorm:"-" temporaljson:"vcs_connection_type,omitzero,omitempty"`

	Variables      pgtype.Hstore  `json:"variables,omitzero" gorm:"type:hstore" swaggertype:"object,string" features:"template" temporaljson:"variables,omitzero,omitempty"`
	EnvVars        pgtype.Hstore  `json:"env_vars,omitzero" temporalsjson:"env_vars" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"env_vars,omitzero,omitempty"`
	VariablesFiles pq.StringArray `gorm:"type:text[]" json:"variables_files,omitzero" swaggertype:"array,string" features:"template" temporaljson:"variables_files,omitzero,omitempty"`

	References pq.StringArray `json:"references" temporaljson:"references" swaggertype:"array,string" gorm:"type:text[]"`
	Refs       []refs.Ref     `gorm:"-"`

	TerraformVersion string `json:"terraform_version,omitzero" gorm:"notnull" temporaljson:"terraform_version,omitzero,omitempty"`
	DriftSchedule    string `json:"drift_schedule,omitzero" gorm:"default null" temporaljson:"drift_schedule,omitzero,omitempty"`

	InstallSandboxRuns []InstallSandboxRun `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_sandbox_runs,omitzero,omitempty"`

	// cloud specific fields
	AWSRegionType generics.NullString `json:"aws_region_type,omitzero" swaggertype:"string" temporaljson:"aws_region_type,omitzero,omitempty"`

	// fields set via after query
	CloudPlatform CloudPlatform `json:"cloud_platform,omitzero" gorm:"-" swaggertype:"string" temporaljson:"cloud_platform,omitzero,omitempty"`
}

func (c *AppSandboxConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppSandboxConfig{}, "preload"),
			Columns: []string{
				"app_id",
				"deleted_at",
				"created_at DESC",
			},
		},
	}
}

// NOTE: currently, only public repo vcs configs are supported when rendering policies and artifacts
func (c *AppSandboxConfig) AfterQuery(tx *gorm.DB) error {
	cRefs := make([]refs.Ref, 0)
	for _, ref := range c.References {
		cRefs = append(cRefs, refs.NewFromString(ref))
	}
	c.Refs = cRefs

	// set the vcs connection type correctly
	if c.ConnectedGithubVCSConfig != nil {
		c.VCSConnectionType = VCSConnectionTypeConnectedRepo
	} else if c.PublicGitVCSConfig != nil {
		c.VCSConnectionType = VCSConnectionTypePublicRepo
	} else {
		c.VCSConnectionType = VCSConnectionTypeNone
	}

	return nil
}

func (a *AppSandboxConfig) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
