package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type AWSAccount struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `faker:"-" json:"-"`

	InstallID string `json:"-" gorm:"notnull"`

	Region     string `json:"region" gorm:"notnull"`
	IAMRoleARN string `json:"iam_role_arn" gorm:"notnull"`
}

func (a *AWSAccount) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewAWSAccountID()
	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
