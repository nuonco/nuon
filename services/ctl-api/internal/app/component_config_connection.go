package app

import (
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/lib/pq"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/viewsql"
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

	AppConfigID string    `json:"app_config_id,omitzero" temporaljson:"app_config_id,omitzero,omitempty"`
	AppConfig   AppConfig `json:"-" temporaljson:"app_config,omitzero,omitempty"`

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
	PulumiComponentConfig             *PulumiComponentConfig             `json:"pulumi,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"pulumi_component_config,omitzero,omitempty"`
	ComponentDependencyIDs            pq.StringArray                     `json:"component_dependency_ids" temporaljson:"component_dependency_ids" swaggertype:"array,string" gorm:"type:text[]"`
	References                        pq.StringArray                     `json:"references" temporaljson:"references" swaggertype:"array,string" gorm:"type:text[]"`
	Checksum                          string                             `json:"checksum,omitzero" gorm:"default null" temporaljson:"checksum,omitzero,omitempty"`
	DriftSchedule                     string                             `json:"drift_schedule,omitzero" gorm:"default null" temporaljson:"drift_schedule,omitzero,omitempty"`
	BuildTimeout                      string                             `json:"build_timeout,omitempty" gorm:"default:null" temporaljson:"build_timeout,omitzero,omitempty"`   // Duration string for build operations (e.g., "30m", "1h"). Max 1h.
	DeployTimeout                     string                             `json:"deploy_timeout,omitempty" gorm:"default:null" temporaljson:"deploy_timeout,omitzero,omitempty"` // Duration string for deploy operations (e.g., "30m", "1h"). Max 1h.
	MaxAutoRetries                    *int                               `json:"max_auto_retries,omitempty" gorm:"default:null" temporaljson:"max_auto_retries,omitzero,omitempty"`
	SkipNoops                         *bool                              `json:"skip_noops,omitempty" gorm:"default:null" temporaljson:"skip_noops,omitzero,omitempty"`
	AutoApproveOnPoliciesPassing      *bool                              `json:"auto_approve_on_policies_passing,omitempty" gorm:"default:null" temporaljson:"auto_approve_on_policies_passing,omitzero,omitempty"`
	Toggleable                        *bool                              `json:"toggleable,omitempty" gorm:"default:null" temporaljson:"toggleable,omitzero,omitempty"`
	DefaultEnabled                    *bool                              `json:"default_enabled,omitempty" gorm:"default:null" temporaljson:"default_enabled,omitzero,omitempty"`

	// Operation roles map: operation type -> role name
	OperationRoles pgtype.Hstore `json:"operation_roles,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"operation_roles,omitzero,omitempty"`

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
		{
			Name: indexes.Name(db, &ComponentConfigConnection{}, "app_config_id_deleted_at"),
			Columns: []string{
				"app_config_id",
				"deleted_at",
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
	if c.PulumiComponentConfig != nil {
		c.Type = ComponentTypePulumi
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
	} else if c.KubernetesManifestComponentConfig != nil {
		c.ConnectedGithubVCSConfig = c.KubernetesManifestComponentConfig.ConnectedGithubVCSConfig
		c.PublicGitVCSConfig = c.KubernetesManifestComponentConfig.PublicGitVCSConfig
	} else if c.PulumiComponentConfig != nil {
		c.ConnectedGithubVCSConfig = c.PulumiComponentConfig.ConnectedGithubVCSConfig
		c.PublicGitVCSConfig = c.PulumiComponentConfig.PublicGitVCSConfig
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
	if c.SkipNoops == nil {
		c.SkipNoops = generics.ToPtr(false)
	}
	return nil
}

func (c *ComponentConfigConnection) GetBuildTimeout() *time.Duration {
	if c.BuildTimeout != "" {
		d, err := time.ParseDuration(c.BuildTimeout)
		if err != nil {
			return nil
		}
		return &d
	}
	return nil
}

func (c *ComponentConfigConnection) GetDeployTimeout() *time.Duration {
	if c.DeployTimeout != "" {
		d, err := time.ParseDuration(c.DeployTimeout)
		if err != nil {
			return nil
		}
		return &d
	}
	return nil
}

func (c *ComponentConfigConnection) GetMaxAutoRetries() int {
	if c.MaxAutoRetries != nil {
		return *c.MaxAutoRetries
	}
	return 0 // default to disabled
}

func (c *ComponentConfigConnection) GetSkipNoops() bool {
	if c.SkipNoops != nil {
		return *c.SkipNoops
	}
	return false // default to not skipping noops — opt-in
}

func (c *ComponentConfigConnection) GetAutoApproveOnPoliciesPassing() bool {
	if c.AutoApproveOnPoliciesPassing != nil {
		return *c.AutoApproveOnPoliciesPassing
	}
	return false // default to not auto-approving — opt-in
}

func (c *ComponentConfigConnection) IsToggleable() bool {
	if c.Toggleable != nil {
		return *c.Toggleable
	}
	return false
}

func (c *ComponentConfigConnection) GetDefaultEnabled() bool {
	if c.DefaultEnabled != nil {
		return *c.DefaultEnabled
	}
	return false
}

// ComponentEnabledFromInputs resolves whether a toggleable component is enabled
// from a set of install input values. The synthetic enabled input
// (config.EnabledOverrideInputName) is the source of truth; when unset it falls
// back to the component's default_enabled. Non-toggleable components are always
// enabled.
func ComponentEnabledFromInputs(enabledInputs map[string]*string, ccc *ComponentConfigConnection) bool {
	if ccc == nil || !ccc.IsToggleable() {
		return true
	}
	name := config.EnabledOverrideInputName(ccc.Component.Name)
	if v, ok := enabledInputs[name]; ok && v != nil {
		if enabled, err := strconv.ParseBool(*v); err == nil {
			return enabled
		}
	}
	return ccc.GetDefaultEnabled()
}
