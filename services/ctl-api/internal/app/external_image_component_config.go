package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type ExternalImageComponentConfig struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// value
	ComponentConfigConnectionID string `json:"component_config_connection_id"`

	// VCS Config
	PublicGitVCSConfig       *PublicGitVCSConfig       `gorm:"polymorphic:ComponentConfig" json:"public_git_vcs_config,omitempty"`
	ConnectedGithubVCSConfig *ConnectedGithubVCSConfig `gorm:"polymorphic:ComponentConfig" json:"connected_github_vcs_config,omitempty"`

	ImageURL          string             `json:"image_url"`
	Tag               string             `json:"tag"`
	AWSECRImageConfig *AWSECRImageConfig `gorm:"polymorphic:ComponentConfig" json:"aws_ecr_image_config,omitempty"`

	SyncOnly          bool               `json:"sync_only,omitempty"`
	BasicDeployConfig *BasicDeployConfig `gorm:"polymorphic:ComponentConfig" json:"basic_deploy_config,omitempty"`
}

func (e *ExternalImageComponentConfig) BeforeCreate(tx *gorm.DB) error {
	e.ID = domains.NewComponentID()
	return nil
}
