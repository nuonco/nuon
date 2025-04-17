package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/iancoleman/strcase"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type AppSecretConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id" gorm:"notnull;default null"`
	Org   Org    `faker:"-" json:"-"`

	AppID string `json:"app_id"`
	App   App    `faker:"-" json:"-"`

	AppConfigID string `json:"app_config_id"`

	AppSecretsConfig   AppSecretsConfig `json:"-" faker:"-"`
	AppSecretsConfigID string           `json:"app_secrets_config_id"`

	Name        string `json:"name" features:"template"`
	DisplayName string `json:"display_name" features:"template"`
	Description string `json:"description" features:"template"`

	Required bool `json:"required"`

	// for syncing into kubernetes
	KubernetesSecretNamespace string `json:"kubernetes_secret_namespace" features:"template"`
	KubernetesSecretName      string `json:"kubernetes_secret_name" features:"template"`

	CloudFormationStackName string `json:"cloudformation_stack_name" gorm:"-"`
	CloudFormationParamName string `json:"cloudformation_param_name" gorm:"-"`
}

func (a *AppSecretConfig) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppCfgID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (a *AppSecretConfig) AfterQuery(tx *gorm.DB) error {
	cfnName := strcase.ToCamel(a.Name)
	a.CloudFormationStackName = cfnName

	a.CloudFormationParamName = cfnName + "Param"
	return nil
}
