package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type ComponentStatus string

const (
	ComponentStatusError          ComponentStatus = "error"
	ComponentStatusActive         ComponentStatus = "active"
	ComponentStatusDeprovisioning ComponentStatus = "deprovisioning"
)

type ComponentType string

const (
	ComponentTypeTerraformModule ComponentType = "terraform_module"
	ComponentTypeHelmChart       ComponentType = "helm_chart"
	ComponentTypeDockerBuild     ComponentType = "docker_build"
	ComponentTypeExternalImage   ComponentType = "external_image"
	ComponentTypeJob             ComponentType = "job"
	ComponentTypeUnknown         ComponentType = "unknown"
)

type Component struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26;" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_app_component_name,unique" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	Name    string `json:"name" gorm:"notnull;index:idx_app_component_name,unique"`
	VarName string `json:"var_name"`

	AppID string `json:"app_id" gorm:"notnull;index:idx_app_component_name,unique"`
	App   App    `faker:"-" json:"-"`

	Status            ComponentStatus `json:"status" swaggertype:"string"`
	StatusDescription string          `json:"status_description"`

	ConfigVersions    int                         `gorm:"-" json:"config_versions"`
	ComponentConfigs  []ComponentConfigConnection `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	InstallComponents []InstallComponent          `gorm:"constraint:OnDelete:CASCADE;" json:"-"`

	Dependencies  []*Component `gorm:"many2many:component_dependencies;constraint:OnDelete:CASCADE;" json:"-"`
	DependencyIDs []string     `gorm:"-" json:"dependencies"`

	// after query loaded items

	Type            ComponentType              `gorm:"-" json:"type"`
	LatestConfig    *ComponentConfigConnection `gorm:"-" json:"-"`
	ResolvedVarName string                     `json:"resolved_var_name" gorm:"-"`
}

func (c *Component) AfterQuery(tx *gorm.DB) error {
	c.ResolvedVarName = generics.First(c.VarName, c.Name)

	// set dependency ids
	for _, dep := range c.Dependencies {
		c.DependencyIDs = append(c.DependencyIDs, dep.ID)
	}

	// set configs
	c.ConfigVersions = len(c.ComponentConfigs)
	c.Type = ComponentTypeUnknown
	if len(c.ComponentConfigs) < 1 {
		return nil
	}

	// parse the latest config
	c.LatestConfig = &c.ComponentConfigs[0]
	if c.LatestConfig.HelmComponentConfig != nil {
		c.Type = ComponentTypeHelmChart
	}
	if c.LatestConfig.TerraformModuleComponentConfig != nil {
		c.Type = ComponentTypeTerraformModule
	}
	if c.LatestConfig.DockerBuildComponentConfig != nil {
		c.Type = ComponentTypeDockerBuild
	}
	if c.LatestConfig.ExternalImageComponentConfig != nil {
		c.Type = ComponentTypeExternalImage
	}
	if c.LatestConfig.JobComponentConfig != nil {
		c.Type = ComponentTypeJob
	}

	return nil
}

func (c *Component) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
