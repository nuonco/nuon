package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/lib/pq"

	"github.com/powertoolsdev/mono/pkg/config/refs"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
)

type VCSConnectionType string

const (
	VCSConnectionTypeConnectedRepo VCSConnectionType = "connected_repo"
	VCSConnectionTypePublicRepo    VCSConnectionType = "public_repo"
	VCSConnectionTypeNone          VCSConnectionType = "none"
)

type ComponentConfigConnection struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	AppConfigID string `json:"app_config_id,omitzero" temporaljson:"app_config_id,omitzero,omitempty"`

	ComponentID   string    `json:"component_id,omitzero" gorm:"notnull" temporaljson:"component_id,omitzero,omitempty"`
	ComponentName string    `json:"component_name,omitzero" gorm:"-" temporaljson:"component_name,omitzero,omitempty"`
	Component     Component `json:"-" temporaljson:"component,omitzero,omitempty"`

	ComponentBuilds []ComponentBuild `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"component_builds,omitzero,omitempty"`

	TerraformModuleComponentConfig    *TerraformModuleComponentConfig    `json:"terraform_module,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"terraform_module_component_config,omitzero,omitempty"`
	HelmComponentConfig               *HelmComponentConfig               `json:"helm,omitempty,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"helm_component_config,omitzero,omitempty"`
	ExternalImageComponentConfig      *ExternalImageComponentConfig      `json:"external_image,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"external_image_component_config,omitzero,omitempty"`
	DockerBuildComponentConfig        *DockerBuildComponentConfig        `json:"docker_build,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"docker_build_component_config,omitzero,omitempty"`
	JobComponentConfig                *JobComponentConfig                `json:"job,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"job_component_config,omitzero,omitempty"`
	KubernetesManifestComponentConfig *KubernetesManifestComponentConfig `json:"kubernetes_manifest,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"kubernetes_manifest_component_config,omitzero,omitempty"`
	ComponentDependencyIDs            pq.StringArray                     `json:"component_dependency_ids" temporaljson:"component_dependency_ids" swaggertype:"array,string" gorm:"type:text[]"`
	References                        pq.StringArray                     `json:"references" temporaljson:"references" swaggertype:"array,string" gorm:"type:text[]"`
	Checksum                          string                             `json:"checksum,omitzero" gorm:"default null" temporaljson:"checksum,omitzero,omitempty"`
	DriftSchedule                     string                             `json:"drift_schedule,omitzero" gorm:"default null" temporaljson:"drift_schedule,omitzero,omitempty"`

	// loaded via after query
	VCSConnectionType        VCSConnectionType         `json:"-" gorm:"-" temporaljson:"vcs_connection_type,omitzero,omitempty"`
	PublicGitVCSConfig       *PublicGitVCSConfig       `gorm:"-" json:"-" temporaljson:"public_git_vcs_config,omitzero,omitempty"`
	ConnectedGithubVCSConfig *ConnectedGithubVCSConfig `gorm:"-" json:"-" temporaljson:"connected_github_vcs_config,omitzero,omitempty"`

	Type ComponentType `gorm:"-" json:"type,omitzero" temporaljson:"type,omitzero,omitempty"`

	Version          int        `json:"version,omitzero" gorm:"->;-:migration" temporaljson:"version,omitzero,omitempty"`
	AppConfigVersion int        `json:"app_config_version,omitzero" gorm:"->;-:migration" temporaljson:"app_config_version,omitzero,omitempty"`
	Refs             []refs.Ref `gorm:"-"`
}

func (c *ComponentConfigConnection) UseView() bool {
	return true
}

func (c *ComponentConfigConnection) ViewVersion() string {
	return "v1"
}

func (c *ComponentConfigConnection) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.DefaultViewName(db, &ComponentConfigConnection{}, 1),
			SQL:           viewsql.ComponentConfigConnectionsV1,
			AlwaysReapply: true,
		},
		{
			Name:          views.CustomViewName(db, &ComponentConfigConnection{}, "latest_configs_view"),
			SQL:           viewsql.LatestComponentConfigConnectionsV1,
			AlwaysReapply: true,
		},
	}
}

func (a *ComponentConfigConnection) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &ComponentConfigConnection{}, "component_id"),
			Columns: []string{
				"component_id",
				"deleted_at",
			},
		},
		{
			Name: indexes.Name(db, &ComponentConfigConnection{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (c *ComponentConfigConnection) AfterQuery(tx *gorm.DB) error {
	cRefs := make([]refs.Ref, 0)
	for _, ref := range c.References {
		cRefs = append(cRefs, refs.NewFromString(ref))
	}
	c.Refs = cRefs

	c.Type = ComponentTypeUnknown
	if c.HelmComponentConfig != nil {
		c.Type = ComponentTypeHelmChart
	}
	if c.TerraformModuleComponentConfig != nil {
		c.Type = ComponentTypeTerraformModule
	}
	if c.DockerBuildComponentConfig != nil {
		c.Type = ComponentTypeDockerBuild
	}
	if c.ExternalImageComponentConfig != nil {
		c.Type = ComponentTypeExternalImage
	}
	if c.JobComponentConfig != nil {
		c.Type = ComponentTypeJob
	}
	if c.KubernetesManifestComponentConfig != nil {
		c.Type = ComponentTypeKubernetesManifest
	}

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

	// set the type

	if c.Component.Name != "" {
		c.ComponentName = c.Component.Name
	}

	return nil
}

func (c *ComponentConfigConnection) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
