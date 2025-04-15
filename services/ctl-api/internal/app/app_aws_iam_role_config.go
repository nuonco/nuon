package app

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/iancoleman/strcase"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type AWSIAMRoleType string

const (
	// used for initial install setup
	AWSIAMRoleTypeRunnerProvision AWSIAMRoleType = "runner_provision"
	// used for tearing down an install
	AWSIAMRoleTypeRunnerDeprovision AWSIAMRoleType = "runner_deprovision"
	// used for updates and other maintenance
	AWSIAMRoleTypeRunnerMaintenance AWSIAMRoleType = "runner_maintenance"

	// used for break-glass by the vendor
	AWSIAMRoleTypeBreakGlass AWSIAMRoleType = "breakglass"

	// used for break glass mode where the runner is given elevated permissions
	//
	// NOTE(jm): at some point, we probably need break glass actions
	AWSIAMRoleTypeRunnerBreakGlass AWSIAMRoleType = "runner_breakglass"
)

type AppAWSIAMRoleConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id" gorm:"notnull;default null"`
	Org   Org    `faker:"-" json:"-"`

	AppConfigID string `json:"app_config_id"`

	Type        AWSIAMRoleType `json:"type"`
	Name        string         `json:"name" features:"template"`
	Description string         `json:"description" features:"template"`
	DisplayName string         `json:"display_name" features:"template"`

	OwnerID   string `json:"owner_id" gorm:"type:text;check:owner_id_checker,char_length(id)=26"`
	OwnerType string `json:"owner_type" gorm:"type:text;"`

	Policies                     []AppAWSIAMPolicyConfig `json:"policies" gorm:"constraint:OnDelete:CASCADE;"`
	PermissionsBoundaryJSON      []byte                  `json:"permissions_boundary" gorm:"type:jsonb" swaggertype:"string" features:"template"`
	CloudFormationStackName      string                  `json:"cloudformation_stack_name" gorm:"-" features:"template"`
	CloudFormationStackParamName string                  `json:"cloudformation_param_name" gorm:"-" features:"template"`
}

func (a *AppAWSIAMRoleConfig) AfterQuery(tx *gorm.DB) error {
	cfnName := strcase.ToCamel(string(a.Type))
	if a.Type == AWSIAMRoleTypeRunnerBreakGlass {
		cfnName = strcase.ToCamel(fmt.Sprintf("%s%s", a.Type, a.Name))
	}

	a.CloudFormationStackName = cfnName
	a.CloudFormationStackParamName = "Enable" + cfnName
	return nil
}

func (a *AppAWSIAMRoleConfig) BeforeCreate(tx *gorm.DB) error {
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
