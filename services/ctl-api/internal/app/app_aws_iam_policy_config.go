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

type AppAWSIAMPolicyConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppConfigID string `json:"app_config_id,omitzero" temporaljson:"app_config_id,omitzero,omitempty"`

	AppAWSIAMRoleConfigID string              `json:"app_aws_iam_role_config_id,omitzero" temporaljson:"app_awsiam_role_config_id,omitzero,omitempty"`
	AppAWSIAMRoleConfig   AppAWSIAMRoleConfig `json:"-" temporaljson:"app_awsiam_role_config,omitzero,omitempty"`

	ManagedPolicyName       string `json:"managed_policy_name,omitzero" features:"template" temporaljson:"managed_policy_name,omitzero,omitempty"`
	Name                    string `json:"name" features:"template,omitzero" temporaljson:"name,omitzero,omitempty"`
	Contents                []byte `json:"contents,omitzero" gorm:"type:jsonb" swaggertype:"string" features:"template" temporaljson:"contents,omitzero,omitempty"`
	CloudFormationStackName string `json:"cloudformation_stack_name,omitzero" gorm:"-" features:"template" temporaljson:"cloud_formation_stack_name,omitzero,omitempty"`
}

func (a *AppAWSIAMPolicyConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppAWSIAMPolicyConfig{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (a *AppAWSIAMPolicyConfig) AfterQuery(tx *gorm.DB) error {
	cfnName := strcase.ToCamel(string(a.Name))
	a.CloudFormationStackName = cfnName

	return nil
}

func (a *AppAWSIAMPolicyConfig) BeforeCreate(tx *gorm.DB) error {
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
