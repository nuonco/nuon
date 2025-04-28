package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type AppRunnerType string

const (
	AppRunnerTypeAWSECS   AppRunnerType = "aws-ecs"
	AppRunnerTypeAWSEKS   AppRunnerType = "aws-eks"
	AppRunnerTypeAzureAKS AppRunnerType = "azure-aks"
	AppRunnerTypeAzureACS AppRunnerType = "azure-acs"
	AppRunnerTypeLocal    AppRunnerType = "local"

	// the aws independent runner
	AppRunnerTypeAWS AppRunnerType = "aws"
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
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID       string `json:"org_id" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org         Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`
	AppID       string `json:"app_id" temporaljson:"app_id,omitzero,omitempty"`
	AppConfigID string `json:"app_config_id" temporaljson:"app_config_id,omitzero,omitempty"`

	EnvVars pgtype.Hstore `json:"env_vars" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"env_vars,omitzero,omitempty"`
	Type    AppRunnerType `json:"app_runner_type" gorm:"not null;default null;" temporaljson:"type,omitzero,omitempty"`

	HelmDriver AppRunnerConfigHelmDriverType `json:"helm_driver" gorm:"default null" swaggertype:"string" temporaljson:"helm_driver,omitzero,omitempty"`
	// ↑ for the runner helm client: only relevant for k8s sandboxes

	// fields set via after query

	CloudPlatform CloudPlatform `json:"cloud_platform" gorm:"-" temporaljson:"cloud_platform,omitzero,omitempty"`
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
	case AppRunnerTypeAzureAKS, AppRunnerTypeAzureACS:
		a.CloudPlatform = CloudPlatformAzure
	default:
		a.CloudPlatform = CloudPlatformUnknown
	}
	return nil
}
