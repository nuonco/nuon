package app

import (
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/iancoleman/strcase"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

var (
	namedPolicyTemplateExpr = regexp.MustCompile(`\{\{[^}]*\}\}`)
	namedPolicyNonAlnum     = regexp.MustCompile(`[^a-zA-Z0-9]+`)
)

// NamedIAMPolicyCloudFormationStackName is the template logical ID for a
// customer-managed IAM policy. The prefix keeps it from colliding with inline
// AWS::IAM::Policy resources, which use ToCamel(policy.Name) alone. Template
// expressions are stripped so "{{.nuon.install.id}}-alb-create" becomes
// NamedPolicyAlbCreate.
func NamedIAMPolicyCloudFormationStackName(name string) string {
	stripped := namedPolicyTemplateExpr.ReplaceAllString(name, " ")
	stripped = namedPolicyNonAlnum.ReplaceAllString(stripped, " ")
	camel := strcase.ToCamel(stripped)
	if camel == "" {
		camel = "Policy"
	}
	return "NamedPolicy" + camel
}

// AppNamedIAMPolicyConfig is an install-stack IAM managed policy defined once
// and attached to any number of roles. It is owned by the permissions config,
// not a role, so CloudFormation can create it even when every role is disabled.
type AppNamedIAMPolicyConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppConfigID string `json:"app_config_id,omitzero" temporaljson:"app_config_id,omitzero,omitempty"`

	AppPermissionsConfigID string               `json:"app_permissions_config_id,omitzero" temporaljson:"app_permissions_config_id,omitzero,omitempty"`
	AppPermissionsConfig   AppPermissionsConfig `json:"-" temporaljson:"app_permissions_config,omitzero,omitempty"`

	// Name is the config identifier and the AWS IAM managed policy name.
	// Roles attach this policy by repeating the same name.
	Name string `json:"name" features:"template,omitzero" temporaljson:"name,omitzero,omitempty"`
	// PolicyName is the AWS IAM managed policy name. Empty means use Name.
	PolicyName  string `json:"policy_name,omitzero" features:"template" temporaljson:"policy_name,omitzero,omitempty"`
	Description string `json:"description,omitzero" features:"template" temporaljson:"description,omitzero,omitempty"`
	Contents    []byte `json:"contents,omitzero" gorm:"type:jsonb" swaggertype:"string" features:"template" temporaljson:"contents,omitzero,omitempty"`

	CloudFormationStackName string `json:"cloudformation_stack_name,omitzero" gorm:"-" features:"template" temporaljson:"cloud_formation_stack_name,omitzero,omitempty"`
}

func (a *AppNamedIAMPolicyConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppNamedIAMPolicyConfig{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &AppNamedIAMPolicyConfig{}, "perms_cfg_id_deleted_at"),
			Columns: []string{
				"app_permissions_config_id",
				"deleted_at",
			},
		},
	}
}

func (a *AppNamedIAMPolicyConfig) AfterQuery(tx *gorm.DB) error {
	a.CloudFormationStackName = NamedIAMPolicyCloudFormationStackName(a.Name)
	return nil
}

func (a *AppNamedIAMPolicyConfig) BeforeCreate(tx *gorm.DB) error {
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

func (a *AppNamedIAMPolicyConfig) AWSPolicyName() string {
	if a.PolicyName != "" {
		return a.PolicyName
	}
	return a.Name
}

// AWSPolicyNameForInstall is the account-global IAM managed policy name.
// IAM policy names are unique per account, so the install id is prefixed
// unless the name is already install-scoped.
func (a *AppNamedIAMPolicyConfig) AWSPolicyNameForInstall(installID string) string {
	name := a.AWSPolicyName()
	if installID == "" || strings.HasPrefix(name, installID+"-") {
		return name
	}
	return installID + "-" + name
}
