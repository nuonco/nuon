package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/iancoleman/strcase"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type AppAWSIAMPolicyConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id" gorm:"notnull;default null"`
	Org   Org    `faker:"-" json:"-"`

	AppConfigID string `json:"app_config_id"`

	AppAWSIAMRoleConfigID string              `json:"app_aws_iam_role_config_id"`
	AppAWSIAMRoleConfig   AppAWSIAMRoleConfig `json:"-"`

	ManagedPolicyName       string `json:"managed_policy_name" features:"template"`
	Name                    string `json:"name" features:"template"`
	Contents                []byte `json:"contents" gorm:"type:jsonb" swaggertype:"string" features:"template"`
	CloudFormationStackName string `json:"cloudformation_stack_name" gorm:"-" features:"template"`
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
