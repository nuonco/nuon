package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type AWSECRImageConfig struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time      `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"notnull"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull"`

	// connection to parent model
	ComponentConfigID   string `json:"component_config_id" gorm:"notnull"`
	ComponentConfigType string `json:"component_config_type" gorm:"notnull"`

	// actual configuration
	IAMRoleARN string `json:"iam_role_arn" gorm:"notnull"`
	AWSRegion  string `json:"aws_region" gorm:"notnull"`
}

func (c *AWSECRImageConfig) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
