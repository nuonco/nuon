package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type InstallStackOutputs struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id" gorm:"notnull;default null"`
	Org   Org    `faker:"-" json:"-"`

	InstallStackID           string              `json:"install_stack_id" gorm:"notnull;default null"`
	InstallStackVersionRunID generics.NullString `json:"install_version_run_id"`

	Data pgtype.Hstore `json:"data" gorm:"type:hstore" swaggertype:"object,string"`

	AWSStackOutputs       *AWSStackOutputs `json:"aws" gorm:"-"`
	VPCID                 string           `json:"vpc_id" gorm:"-"`
	RunnerSubnet          string           `json:"runner_subnet" gorm:"-"`
	PublicSubnets         []string         `json:"public_subnets" gorm:"-"`
	PrivateSubnets        []string         `json:"private_subnets" gorm:"-"`
	ProvisionIAMRoleARN   string           `json:"provision_iam_role_arn" gorm:"-"`
	DeprovisionIAMRoleARN string           `json:"deprovision_iam_role_arn" gorm:"-"`
	MaintenanceIAMRoleARN string           `json:"maintenance_iam_role_arn" gorm:"-"`
	RunnerIAMRoleARN      string           `json:"runner_iam_role_arn" gorm:"-"`
}

type AWSStackOutputs struct {
	VPCID                 string   `json:"vpc_id" mapstructure:"vpc_id"`
	RunnerSubnet          string   `json:"runner_subnet" mapstructure:"runner_subnet"`
	PublicSubnets         []string `json:"public_subnets" mapstructure:"public_subnets"`
	PrivateSubnets        []string `json:"private_subnets" mapstructure:"private_subnets"`
	ProvisionIAMRoleARN   string   `json:"provision_iam_role_arn" mapstructure:"provision_iam_role_arn"`
	DeprovisionIAMRoleARN string   `json:"deprovision_iam_role_arn" mapstructure:"deprovision_iam_role_arn"`
	MaintenanceIAMRoleARN string   `json:"maintenance_iam_role_arn" mapstructure:"maintenance_iam_role_arn"`
	RunnerIAMRoleARN      string   `json:"runner_iam_role_arn" mapstructure:"runner_iam_role_arn"`
}

func (a *InstallStackOutputs) BeforeCreate(tx *gorm.DB) error {
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
