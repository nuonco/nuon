package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type TerraformModuleComponentConfig struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time      `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"notnull"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// parent reference
	ComponentConfigConnectionID string `json:"component_config_connection_id" gorm:"notnull"`

	// terraform configuration values
	Version   string        `json:"version" gorm:"default:v1.5.3;notnull"`
	Variables pgtype.Hstore `json:"variables" gorm:"type:hstore" swaggertype:"object,string"`

	// VCSConfig
	PublicGitVCSConfig       *PublicGitVCSConfig       `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"public_git_vcs_config,omitempty"`
	ConnectedGithubVCSConfig *ConnectedGithubVCSConfig `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"connected_github_vcs_config,omitempty"`
}

func (c *TerraformModuleComponentConfig) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	return nil
}
