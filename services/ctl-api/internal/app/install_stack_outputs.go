package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type InstallStackOutputs struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	InstallStackID           string              `json:"install_stack_id,omitzero" gorm:"notnull;default null" temporaljson:"install_stack_id,omitzero,omitempty"`
	InstallStackVersionRunID generics.NullString `json:"install_version_run_id,omitzero" swaggertype:"string" temporaljson:"install_stack_version_run_id,omitzero,omitempty"`

	Data pgtype.Hstore `json:"data,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"data,omitzero,omitempty"`

	AWSStackOutputs *AWSStackOutputs `json:"aws,omitzero" gorm:"-" temporaljson:"aws_stack_outputs,omitzero,omitempty"`
}

type AWSStackOutputs struct {
	AccountID             string   `json:"account_id,omitzero" mapstructure:"account_id" temporaljson:"account_id,omitzero,omitempty"`
	Region                string   `json:"region,omitzero" mapstructure:"region" temporaljson:"region,omitzero,omitempty"`
	VPCID                 string   `json:"vpc_id,omitzero" mapstructure:"vpc_id" temporaljson:"vpcid,omitzero,omitempty"`
	RunnerSubnet          string   `json:"runner_subnet,omitzero" mapstructure:"runner_subnet" temporaljson:"runner_subnet,omitzero,omitempty"`
	PublicSubnets         []string `json:"public_subnets,omitzero" mapstructure:"public_subnets" temporaljson:"public_subnets,omitzero,omitempty"`
	PrivateSubnets        []string `json:"private_subnets,omitzero" mapstructure:"private_subnets" temporaljson:"private_subnets,omitzero,omitempty"`
	ProvisionIAMRoleARN   string   `json:"provision_iam_role_arn,omitzero" mapstructure:"provision_iam_role_arn" temporaljson:"provision_iam_role_arn,omitzero,omitempty"`
	DeprovisionIAMRoleARN string   `json:"deprovision_iam_role_arn,omitzero" mapstructure:"deprovision_iam_role_arn" temporaljson:"deprovision_iam_role_arn,omitzero,omitempty"`
	MaintenanceIAMRoleARN string   `json:"maintenance_iam_role_arn,omitzero" mapstructure:"maintenance_iam_role_arn" temporaljson:"maintenance_iam_role_arn,omitzero,omitempty"`
	RunnerIAMRoleARN      string   `json:"runner_iam_role_arn,omitzero" mapstructure:"runner_iam_role_arn" temporaljson:"runner_iam_role_arn,omitzero,omitempty"`
}

func (a *InstallStackOutputs) AfterQuery(tx *gorm.DB) error {
	if len(a.Data) < 1 {
		return nil
	}

	var outputs AWSStackOutputs
	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.StringToTimeDurationHookFunc(),
		),
		WeaklyTypedInput: true,
		Result:           &outputs,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create decoder")
	}
	if err := decoder.Decode(a.Data); err != nil {
		return errors.Wrap(err, "unable to parse aws outputs")
	}
	a.AWSStackOutputs = &outputs

	return nil
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
