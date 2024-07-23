package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type AppAWSDelegationConfig struct {
	ID          string                `gorm:"primarykey" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_aws_delegation_config,unique"`

	OrgID string `json:"org_id" gorm:"notnull;default null"`
	Org   Org    `faker:"-" json:"-"`

	IAMRoleARN string `json:"iam_role_arn"`

	AppSandboxConfigID string `json:"app_sandbox_config_id" gorm:"index:idx_aws_delegation_config,unique"`

	// static credentials for long lived cross account access.
	// NOTE: this is not recommended for long-term usage, just to be used for short term access before gov-cloud
	// support is fully spun up.
	AccessKeyID     string `temporaljson:"access_key_id" json:"-"`
	SecretAccessKey string `temporaljson:"secret_access_key" json:"-"`
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
