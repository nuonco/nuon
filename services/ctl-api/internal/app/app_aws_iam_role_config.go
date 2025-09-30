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
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppConfigID string `json:"app_config_id,omitzero" temporaljson:"app_config_id,omitzero,omitempty"`

	Type        AWSIAMRoleType `json:"type,omitzero" temporaljson:"type,omitzero,omitempty"`
	Name        string         `json:"name,omitzero" features:"template" temporaljson:"name,omitzero,omitempty"`
	Description string         `json:"description,omitzero" features:"template" temporaljson:"description,omitzero,omitempty"`
	DisplayName string         `json:"display_name,omitzero" features:"template" temporaljson:"display_name,omitzero,omitempty"`

	OwnerID   string `json:"owner_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26" temporaljson:"owner_id,omitzero,omitempty"`
	OwnerType string `json:"owner_type,omitzero" gorm:"type:text;" temporaljson:"owner_type,omitzero,omitempty"`

	Policies                     []AppAWSIAMPolicyConfig `json:"policies,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"policies,omitzero,omitempty"`
	PermissionsBoundaryJSON      []byte                  `json:"permissions_boundary,omitzero" gorm:"type:jsonb" swaggertype:"string" features:"template" temporaljson:"permissions_boundary_json,omitzero,omitempty"`
	CloudFormationStackName      string                  `json:"cloudformation_stack_name,omitzero" gorm:"-" features:"template" temporaljson:"cloud_formation_stack_name,omitzero,omitempty"`
	CloudFormationStackParamName string                  `json:"cloudformation_param_name,omitzero" gorm:"-" features:"template" temporaljson:"cloud_formation_stack_param_name,omitzero,omitempty"`
}

func (a *AppAWSIAMRoleConfig) AfterQuery(tx *gorm.DB) error {
	cfnName := strcase.ToCamel(string(a.Type))
	if a.Type == AWSIAMRoleTypeRunnerBreakGlass || a.Type == AWSIAMRoleTypeBreakGlass {
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
