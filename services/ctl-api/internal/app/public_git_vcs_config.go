package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type PublicGitVCSConfig struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	ComponentConfigID   string `json:"component_config_id" gorm:"notnull"`
	ComponentConfigType string `json:"component_config_type" gorm:"notnull"`

	// actual configuration
	Repo      string `json:"repo" gorm:"notnull"`
	Directory string `json:"directory" gorm:"notnull"`
	Branch    string `json:"branch" gorm:"notnull"`
}

func (c *PublicGitVCSConfig) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewVCSID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
