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

type AppRunnerConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	OrgID       string `json:"org_id" gorm:"notnull;default null"`
	Org         Org    `faker:"-" json:"-"`
	AppID       string `json:"app_id"`
	AppConfigID string `json:"app_config_id"`

	EnvVars pgtype.Hstore `json:"env_vars" gorm:"type:hstore" swaggertype:"object,string"`
	Type    AppRunnerType `json:"app_runner_type" gorm:"not null;default null;"`

	// fields set via after query

	CloudPlatform CloudPlatform `json:"cloud_platform" gorm:"-"`
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
