package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/iancoleman/strcase"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type AppSecretConfigFmt string

const (
	AppSecretConfigFmtBase64  AppSecretConfigFmt = "base64"
	AppSecretConfigFmtDefault AppSecretConfigFmt = ""
)

type AppSecretConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppID string `json:"app_id,omitzero" temporaljson:"app_id,omitzero,omitempty"`
	App   App    `faker:"-" json:"-" temporaljson:"app,omitzero,omitempty"`

	AppConfigID string `json:"app_config_id,omitzero" temporaljson:"app_config_id,omitzero,omitempty"`

	AppSecretsConfig   AppSecretsConfig `json:"-" faker:"-" temporaljson:"app_secrets_config,omitzero,omitempty"`
	AppSecretsConfigID string           `json:"app_secrets_config_id,omitzero" temporaljson:"app_secrets_config_id,omitzero,omitempty"`

	Name        string `json:"name,omitzero" features:"template" temporaljson:"name,omitzero,omitempty"`
	DisplayName string `json:"display_name,omitzero" features:"template" temporaljson:"display_name,omitzero,omitempty"`
	Description string `json:"description,omitzero" features:"template" temporaljson:"description,omitzero,omitempty"`

	Required     bool               `json:"required,omitzero" temporaljson:"required,omitzero,omitempty"`
	AutoGenerate bool               `json:"auto_generate,omitzero" temporaljson:"auto_generate,omitzero,omitempty"`
	Format       AppSecretConfigFmt `json:"format" temporaljson:"format" swaggertype:"string"`
	Default      string             `json:"default" temporaljson:"default"`

	// for syncing into kubernetes
	KubernetesSync            bool   `json:"kubernetes_sync,omitzero" temporaljson:"kubernetes_sync,omitzero,omitempty"`
	KubernetesSecretNamespace string `json:"kubernetes_secret_namespace,omitzero" features:"template" temporaljson:"kubernetes_secret_namespace,omitzero,omitempty"`
	KubernetesSecretName      string `json:"kubernetes_secret_name,omitzero" features:"template" temporaljson:"kubernetes_secret_name,omitzero,omitempty"`
	KubernetesSecretKey       string `json:"-" features:"-" temporaljson:"kubernetes_secret_key,omitzero,omitempty"`

	CloudFormationStackName string `json:"cloudformation_stack_name,omitzero" gorm:"-" temporaljson:"cloud_formation_stack_name,omitzero,omitempty"`
	CloudFormationParamName string `json:"cloudformation_param_name,omitzero" gorm:"-" temporaljson:"cloud_formation_param_name,omitzero,omitempty"`
}

func (a *AppSecretConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppSecretConfig{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
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
	a.KubernetesSecretKey = "value"
	return nil
}
