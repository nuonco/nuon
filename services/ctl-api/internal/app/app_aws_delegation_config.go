package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type AppAWSDelegationConfig struct {
	ID          string                `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_aws_delegation_config,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	IAMRoleARN string `json:"iam_role_arn,omitzero" temporaljson:"iam_role_arn,omitzero,omitempty"`

	AppSandboxConfigID string `json:"app_sandbox_config_id,omitzero" gorm:"index:idx_aws_delegation_config,unique" temporaljson:"app_sandbox_config_id,omitzero,omitempty"`

	// static credentials for long lived cross account access.
	// NOTE: this is not recommended for long-term usage, just to be used for short term access before gov-cloud
	// support is fully spun up.
	AccessKeyID     string `json:"-" temporaljson:"access_key_id,omitzero,omitempty"`
	SecretAccessKey string `json:"-" temporaljson:"secret_access_key,omitzero,omitempty"`
}

func (a *AppAWSDelegationConfig) BeforeCreate(tx *gorm.DB) error {
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

func (a *AppAWSDelegationConfig) AfterQuery(tx *gorm.DB) error {
	return nil
}
