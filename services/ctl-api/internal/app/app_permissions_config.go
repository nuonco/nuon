package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type AppPermissionsConfig struct {
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

	Roles []AppAWSIAMRoleConfig `json:"aws_iam_roles" gorm:"constraint:OnDelete:CASCADE;polymorphic:Owner" temporaljson:"roles,omitzero,omitempty"`

	// loaded via an after query
	ProvisionRole   AppAWSIAMRoleConfig `json:"provision_aws_iam_role" gorm:"-" temporaljson:"provision_role,omitzero,omitempty"`
	MaintenanceRole AppAWSIAMRoleConfig `json:"maintenance_aws_iam_role" gorm:"-" temporaljson:"maintenance_role,omitzero,omitempty"`
	DeprovisionRole AppAWSIAMRoleConfig `json:"deprovision_aws_iam_role" gorm:"-" temporaljson:"deprovision_role,omitzero,omitempty"`
}

func (a *AppPermissionsConfig) BeforeCreate(tx *gorm.DB) error {
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

func (a *AppPermissionsConfig) AfterQuery(tx *gorm.DB) error {
	for _, role := range a.Roles {
		switch role.Type {
		case AWSIAMRoleTypeRunnerDeprovision:
			a.DeprovisionRole = role
		case AWSIAMRoleTypeRunnerProvision:
			a.ProvisionRole = role
		case AWSIAMRoleTypeRunnerMaintenance:
			a.MaintenanceRole = role
		default:
		}
	}

	return nil
}
