package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type CloudPlatform string

const (
	CloudPlatformAWS     CloudPlatform = "aws"
	CloudPlatformAzure   CloudPlatform = "azure"
	CloudPlatformUnknown CloudPlatform = "unknown"
)

type AppRunnerType string

const (
	AppRunnerTypeAWSECS   AppRunnerType = "aws-ecs"
	AppRunnerTypeAWSEKS   AppRunnerType = "aws-eks"
	AppRunnerTypeAzureAKS AppRunnerType = "azure-aks"
	AppRunnerTypeAzureACS AppRunnerType = "azure-acs"
)

type AppRunnerConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   UserToken             `json:"created_by" gorm:"references:Subject"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id" gorm:"notnull;default null"`
	Org   Org    `faker:"-" json:"-"`
	AppID string `json:"app_id"`

	EnvVars pgtype.Hstore `json:"env_vars" gorm:"type:hstore" swaggertype:"object,string"`
	Type    AppRunnerType `json:"app_runner_type" gorm:"not null;default null;"`

	// fields set via after query

	CloudPlatform CloudPlatform `json:"cloud_platform" gorm:"-"`
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
	case AppRunnerTypeAWSECS, AppRunnerTypeAWSEKS:
		a.CloudPlatform = CloudPlatformAWS
	case AppRunnerTypeAzureAKS, AppRunnerTypeAzureACS:
		a.CloudPlatform = CloudPlatformAzure
	default:
		a.CloudPlatform = CloudPlatformUnknown
	}
	return nil
}
