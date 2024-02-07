package app

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

const (
	githubRawURLTemplate string = "https://raw.githubusercontent.com/%s/%s/%s/artifacts"
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

	// Links are dynamically loaded using an after query
	Artifacts struct {
		DeprovisionPolicy           string `json:"deprovision_policy" gorm:"-"`
		ProvisionPolicy             string `json:"provision_policy" gorm:"-"`
		TrustPolicy                 string `json:"trust_policy" gorm:"-"`
		CloudformationStackTemplate string `json:"cloudformation_stack_template" gorm:"-"`
	} `json:"artifacts" gorm:"-"`
}

// NOTE: currently, only public repo vcs configs are supported when rendering policies and artifacts
func (c *AppSandboxConfig) AfterQuery(tx *gorm.DB) error {
	vcsCfg := c.PublicGitVCSConfig
	if vcsCfg == nil {
		return nil
	}

	baseURL := fmt.Sprintf(githubRawURLTemplate, vcsCfg.Repo, vcsCfg.Branch, vcsCfg.Directory)
	c.Artifacts.DeprovisionPolicy = fmt.Sprintf("%s/deprovision.json", baseURL)
	c.Artifacts.ProvisionPolicy = fmt.Sprintf("%s/provision.json", baseURL)
	c.Artifacts.TrustPolicy = fmt.Sprintf("%s/trust.json", baseURL)
	c.Artifacts.CloudformationStackTemplate = fmt.Sprintf("%s/template.yaml", baseURL)

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
