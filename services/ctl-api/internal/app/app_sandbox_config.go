package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type AppSandboxConfig struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id" gorm:"notnull;default null"`

	// TODO(jm): add this back, once we have migrated all existing app sandbox configs
	// `gorm:"not null;default null"`
	AppID string `json:"app_id"`

	// NOTE(jm): you can use one of a few different methods of creating an app sandbox, either a built in one, that
	// Nuon manages, or one of the public git vcs configs.

	SandboxReleaseID *string         `json:"sandbox_release_id,omitempty" gorm:"default null"`
	SandboxRelease   *SandboxRelease `json:"sandbox_release,omitempty"`

	// Either a public git repo or private repo using a connected repo source can be used. For now, these fields are
	// not being respected down stream, but will in the future.

	PublicGitVCSConfig       *PublicGitVCSConfig       `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"public_git_vcs_config,omitempty"`
	ConnectedGithubVCSConfig *ConnectedGithubVCSConfig `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"connected_github_vcs_config,omitempty"`

	Variables pgtype.Hstore `json:"variables" gorm:"type:hstore" swaggertype:"object,string"`

	TerraformVersion   string              `json:"terraform_version" gorm:"notnull"`
	InstallSandboxRuns []InstallSandboxRun `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
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
