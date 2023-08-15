package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type AWSAccount struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	InstallID string `json:"-"`

	Region     string `json:"region"`
	IAMRoleARN string `json:"iam_role_arn"`
}

func (a *AWSAccount) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewAWSAccountID()
	return nil
}
