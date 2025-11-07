package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
)

type AppRunnerType string

const (
	AppRunnerTypeUnknown AppRunnerType = "unknown"
	// legacy types from before independent runners
	AppRunnerTypeAWSECS   AppRunnerType = "aws-ecs"
	AppRunnerTypeAWSEKS   AppRunnerType = "aws-eks"
	AppRunnerTypeAzureAKS AppRunnerType = "azure-aks"
	AppRunnerTypeAzureACS AppRunnerType = "azure-acs"

	AppRunnerTypeLocal AppRunnerType = "local"

	// for independent runners
	AppRunnerTypeAWS   AppRunnerType = "aws"
	AppRunnerTypeAzure AppRunnerType = "azure"
)

func (a AppRunnerType) JobType() RunnerJobType {
	switch a {
	case AppRunnerTypeAWSECS, AppRunnerTypeAzureACS:
		return RunnerJobTypeRunnerTerraform
	case AppRunnerTypeAWSEKS, AppRunnerTypeAzureAKS:
		return RunnerJobTypeRunnerHelm
	case AppRunnerTypeLocal, AppRunnerTypeAWS:
		return RunnerJobTypeRunnerLocal
	default:
	}

	return RunnerJobTypeUnknown
}

type AppRunnerConfigHelmDriverType string

const (
	AppRunnerHelmDriverSecret    AppRunnerConfigHelmDriverType = "secret"
	AppRunnerHelmDriverConfigMap AppRunnerConfigHelmDriverType = "configmap"
	AppRunnerHelmDriverEmpty     AppRunnerConfigHelmDriverType = ""
	// ↑ Necessary for records created before this addition
)

type AppRunnerConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID       string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org         Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`
	AppID       string `json:"app_id,omitzero" temporaljson:"app_id,omitzero,omitempty"`
	AppConfigID string `json:"app_config_id,omitzero" temporaljson:"app_config_id,omitzero,omitempty"`

	EnvVars pgtype.Hstore `json:"env_vars,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"env_vars,omitzero,omitempty"`
	Type    AppRunnerType `json:"app_runner_type,omitzero" gorm:"not null;default null;" temporaljson:"type,omitzero,omitempty"`

	HelmDriver AppRunnerConfigHelmDriverType `json:"helm_driver" gorm:"default null" swaggertype:"string" temporaljson:"helm_driver,omitzero,omitempty"`
	// ↑ for the runner helm client: only relevant for k8s sandboxes

	// fields set via after query

	CloudPlatform CloudPlatform `json:"cloud_platform,omitzero" gorm:"-" temporaljson:"cloud_platform,omitzero,omitempty"`

	// takes a URL to a bash script ⤵  which will be `curl | bash`-ed on the VM. usually via user-data or equivalent.
	InitScriptURL string `json:"init_script,omitzero" gorm:"default null" temporaljson:"init_script,omitzero,omitempty" features:"get,omitzero"`
}

func (a *AppRunnerConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppRunnerConfig{}, "preload"),
			Columns: []string{
				"app_id",
				"deleted_at",
				"created_at DESC",
			},
		},
		{
			Name: indexes.Name(db, &AppRunnerConfig{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (i *AppRunnerConfig) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.CustomViewName(db, &AppRunnerConfig{}, "latest_view_v1"),
			SQL:           viewsql.AppRunnerConfigsLatestViewV1,
			AlwaysReapply: true,
		},
	}
}

func (a *AppRunnerConfig) BeforeCreate(tx *gorm.DB) error {
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

func (a *AppRunnerConfig) AfterQuery(tx *gorm.DB) error {
	switch a.Type {
	case AppRunnerTypeAWSECS, AppRunnerTypeAWSEKS, AppRunnerTypeAWS:
		a.CloudPlatform = CloudPlatformAWS
	case AppRunnerTypeAzureAKS, AppRunnerTypeAzureACS, AppRunnerTypeAzure:
		a.CloudPlatform = CloudPlatformAzure
	default:
		a.CloudPlatform = CloudPlatformUnknown
	}

	// configured in the stack generation
	// // NOTE(fd): we set default init scripts here
	// // TODO(fd): use config values
	// if a.InitScript == "" {
	// 	switch a.CloudPlatform {
	// 	case CloudPlatformAWS:
	// 		a.InitScript = "https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/aws/init.sh"
	// 	case CloudPlatformAzure:
	// 		a.InitScript = "https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/azure/init.sh"
	// 	}
	// }
	return nil
}
