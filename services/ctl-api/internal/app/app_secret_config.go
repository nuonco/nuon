package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/iancoleman/strcase"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type AppSecretConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppID string `json:"app_id" temporaljson:"app_id,omitzero,omitempty"`
	App   App    `faker:"-" json:"-" temporaljson:"app,omitzero,omitempty"`

	AppConfigID string `json:"app_config_id" temporaljson:"app_config_id,omitzero,omitempty"`

	AppSecretsConfig   AppSecretsConfig `json:"-" faker:"-" temporaljson:"app_secrets_config,omitzero,omitempty"`
	AppSecretsConfigID string           `json:"app_secrets_config_id" temporaljson:"app_secrets_config_id,omitzero,omitempty"`

	Name        string `json:"name" features:"template" temporaljson:"name,omitzero,omitempty"`
	DisplayName string `json:"display_name" features:"template" temporaljson:"display_name,omitzero,omitempty"`
	Description string `json:"description" features:"template" temporaljson:"description,omitzero,omitempty"`

	Required bool `json:"required" temporaljson:"required,omitzero,omitempty"`

	// for syncing into kubernetes
	KubernetesSecretNamespace string `json:"kubernetes_secret_namespace" features:"template" temporaljson:"kubernetes_secret_namespace,omitzero,omitempty"`
	KubernetesSecretName      string `json:"kubernetes_secret_name" features:"template" temporaljson:"kubernetes_secret_name,omitzero,omitempty"`

	CloudFormationStackName string `json:"cloudformation_stack_name" gorm:"-" temporaljson:"cloud_formation_stack_name,omitzero,omitempty"`
	CloudFormationParamName string `json:"cloudformation_param_name" gorm:"-" temporaljson:"cloud_formation_param_name,omitzero,omitempty"`
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
