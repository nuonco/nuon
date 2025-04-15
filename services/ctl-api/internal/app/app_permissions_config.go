package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type AppPermissionsConfig struct {
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

	Roles []AppAWSIAMRoleConfig `json:"aws_iam_roles" gorm:"constraint:OnDelete:CASCADE;polymorphic:Owner"`

	// loaded via an after query
	ProvisionRole   AppAWSIAMRoleConfig `json:"provision_aws_iam_role" gorm:"-"`
	MaintenanceRole AppAWSIAMRoleConfig `json:"maintenance_aws_iam_role" gorm:"-"`
	DeprovisionRole AppAWSIAMRoleConfig `json:"deprovision_aws_iam_role" gorm:"-"`
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
