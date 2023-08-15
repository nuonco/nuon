package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type AWSECRImageConfig struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// connection to parent model
	ComponentConfigID   string `json:"component_config_id"`
	ComponentConfigType string `json:"component_config_type"`

	// actual configuration
	IAMRoleARN string `json:"iam_role_arn"`
	AWSRegion  string `json:"aws_region"`
}

func (c *AWSECRImageConfig) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	return nil
}
