package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type VCSConnectionType string

const (
	VCSConnectionTypeConnectedRepo VCSConnectionType = "connected_repo"
	VCSConnectionTypePublicRepo    VCSConnectionType = "public_repo"
	VCSConnectionTypeNone          VCSConnectionType = "none"
)

type ComponentConfigConnection struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`

	ComponentID string    `json:"component_id" gorm:"notnull"`
	Component   Component `json:"-"`

	ComponentBuilds []ComponentBuild `json:"-" gorm:"constraint:OnDelete:CASCADE;"`

	TerraformModuleComponentConfig *TerraformModuleComponentConfig `json:"terraform_module,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	HelmComponentConfig            *HelmComponentConfig            `json:"helm,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	ExternalImageComponentConfig   *ExternalImageComponentConfig   `json:"external_image,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	DockerBuildComponentConfig     *DockerBuildComponentConfig     `json:"docker_build,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	JobComponentConfig             *JobComponentConfig             `json:"job,omitempty" gorm:"constraint:OnDelete:CASCADE;"`

	// loaded via after query
	VCSConnectionType        VCSConnectionType         `json:"-" gorm:"-"`
	PublicGitVCSConfig       *PublicGitVCSConfig       `gorm:"-" json:"-"`
	ConnectedGithubVCSConfig *ConnectedGithubVCSConfig `gorm:"-" json:"-"`
}

func (c *ComponentConfigConnection) AfterQuery(tx *gorm.DB) error {
	// set the vcs connection type, by parsing the subfields on the relationship
	if c.TerraformModuleComponentConfig != nil {
		c.ConnectedGithubVCSConfig = c.TerraformModuleComponentConfig.ConnectedGithubVCSConfig
		c.PublicGitVCSConfig = c.TerraformModuleComponentConfig.PublicGitVCSConfig
	} else if c.HelmComponentConfig != nil {
		c.ConnectedGithubVCSConfig = c.HelmComponentConfig.ConnectedGithubVCSConfig
		c.PublicGitVCSConfig = c.HelmComponentConfig.PublicGitVCSConfig
	} else if c.DockerBuildComponentConfig != nil {
		c.ConnectedGithubVCSConfig = c.DockerBuildComponentConfig.ConnectedGithubVCSConfig
		c.PublicGitVCSConfig = c.DockerBuildComponentConfig.PublicGitVCSConfig
	}

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

func (c *ComponentConfigConnection) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
