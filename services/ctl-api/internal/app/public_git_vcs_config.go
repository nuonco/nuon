package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type PublicGitVCSConfig struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	ComponentConfigID   string `json:"component_config_id"`
	ComponentConfigType string `json:"component_config_type"`

	// actual configuration
	Repo      string `json:"repo"`
	Directory string `json:"directory"`
	Branch    string `json:"branch"`
}

func (c *PublicGitVCSConfig) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewVCSID()
	return nil
}
