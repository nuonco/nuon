package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type ExternalImageComponentConfig struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`

	// value
	ComponentConfigConnectionID string `json:"component_config_connection_id" gorm:"notnull"`

	// VCS Config
	PublicGitVCSConfig       *PublicGitVCSConfig       `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"public_git_vcs_config,omitempty"`
	ConnectedGithubVCSConfig *ConnectedGithubVCSConfig `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"connected_github_vcs_config,omitempty"`

	ImageURL          string             `json:"image_url" gorm:"notnull"`
	Tag               string             `json:"tag" gorm:"notnull"`
	AWSECRImageConfig *AWSECRImageConfig `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"aws_ecr_image_config,omitempty"`

	SyncOnly          bool               `json:"sync_only,omitempty"`
	BasicDeployConfig *BasicDeployConfig `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"basic_deploy_config,omitempty"`
}

func (e *ExternalImageComponentConfig) BeforeCreate(tx *gorm.DB) error {
	e.ID = domains.NewComponentID()
	e.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	e.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
