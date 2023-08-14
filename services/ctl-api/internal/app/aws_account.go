package app

import (
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type AWSAccount struct {
	Model

	InstallID string `json:"-"`

	Region     string `json:"region"`
	IAMRoleARN string `json:"iam_role_arn"`
}

func (a *AWSAccount) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewAWSAccountID()
	return nil
}
