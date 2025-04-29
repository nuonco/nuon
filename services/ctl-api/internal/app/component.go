package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/links"
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

func (c ComponentType) SyncJobType() RunnerJobType {
	switch c {
	case ComponentTypeTerraformModule,
		ComponentTypeDockerBuild,
		ComponentTypeExternalImage,
		ComponentTypeHelmChart:
		return RunnerJobTypeOCISync

	case ComponentTypeJob:
		return RunnerJobTypeNOOPBuild
	default:
	}

	return RunnerJobTypeUnknown
}

func (c ComponentType) DeployJobType() RunnerJobType {
	switch c {
	case ComponentTypeTerraformModule:
		return RunnerJobTypeTerraformDeploy
	case ComponentTypeHelmChart:
		return RunnerJobTypeHelmChartDeploy
	case ComponentTypeJob:
		return RunnerJobTypeJobDeploy

		// the following do not require deploys
	case ComponentTypeDockerBuild,
		ComponentTypeExternalImage:
		return RunnerJobTypeJobNOOPDeploy
	default:
	}

	return RunnerJobTypeUnknown
}

func (c ComponentType) BuildJobType() RunnerJobType {
	switch c {
	case ComponentTypeTerraformModule:
		return RunnerJobTypeTerraformModuleBuild
	case ComponentTypeHelmChart:
		return RunnerJobTypeHelmChartBuild
	case ComponentTypeDockerBuild:
		return RunnerJobTypeDockerBuild
	case ComponentTypeExternalImage:
		return RunnerJobTypeContainerImageBuild
	case ComponentTypeJob:
		return RunnerJobTypeNOOPBuild
	default:
	}

	return RunnerJobTypeUnknown
}

type Component struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26;" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_app_component_name,unique" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	Name    string `json:"name" gorm:"notnull;index:idx_app_component_name,unique" temporaljson:"name,omitzero,omitempty"`
	VarName string `json:"var_name" temporaljson:"var_name,omitzero,omitempty"`

	AppID string `json:"app_id" gorm:"notnull;index:idx_app_component_name,unique" temporaljson:"app_id,omitzero,omitempty"`
	App   App    `faker:"-" json:"-" temporaljson:"app,omitzero,omitempty"`

	Status            ComponentStatus `json:"status" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string          `json:"status_description" temporaljson:"status_description,omitzero,omitempty"`

	ConfigVersions    int                         `gorm:"-" json:"config_versions" temporaljson:"config_versions,omitzero,omitempty"`
	ComponentConfigs  []ComponentConfigConnection `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"component_configs,omitzero,omitempty"`
	InstallComponents []InstallComponent          `gorm:"constraint:OnDelete:CASCADE;" json:"-" temporaljson:"install_components,omitzero,omitempty"`

	Dependencies  []*Component `gorm:"many2many:component_dependencies;constraint:OnDelete:CASCADE;" json:"-" temporaljson:"dependencies,omitzero,omitempty"`
	DependencyIDs []string     `gorm:"-" json:"dependencies" temporaljson:"dependency_i_ds,omitzero,omitempty"`

	// after query loaded items

	Links map[string]any `json:"links,omitempty" temporaljson:"-" gorm:"-"`

	Type            ComponentType              `gorm:"-" json:"type" temporaljson:"type,omitzero,omitempty"`
	LatestConfig    *ComponentConfigConnection `gorm:"-" json:"-" temporaljson:"latest_config,omitzero,omitempty"`
	ResolvedVarName string                     `json:"resolved_var_name" gorm:"-" temporaljson:"resolved_var_name,omitzero,omitempty"`
}

func (c *Component) AfterQuery(tx *gorm.DB) error {
	cfg := configFromContext(tx.Statement.Context)
	if cfg != nil {
		c.Links = links.ComponentLinks(cfg, c.ID)
	}

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
	c.Type = c.LatestConfig.Type

	return nil
}

func (c *Component) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (c *Component) JoinTables() []migrations.JoinTable {
	return []migrations.JoinTable{
		{
			Field:     "Dependencies",
			JoinTable: &ComponentDependency{},
		},
	}
}
