package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type VCSConnectionCommit struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	VCSConnection   VCSConnection `json:"-"`
	VCSConnectionID string        `json:"component_config_connection_id"`

	SHA         string `json:"sha"`
	AuthorName  string `json:"author_name"`
	AuthorEmail string `json:"author_email"`
	Message     string `json:"message"`
}

func (v *VCSConnectionCommit) BeforeCreate(tx *gorm.DB) error {
	v.ID = domains.NewVCSConnectionID()
	return nil
}
