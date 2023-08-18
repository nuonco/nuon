package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type Org struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Name   string `gorm:"uniqueIndex" json:"name"`
	Status string `json:"status"`

	Apps           []App           `faker:"-" swaggerignore:"true" json:"apps,omitempty"`
	VCSConnections []VCSConnection `json:"vcs_connections,omitempty"`
	UserOrgs       []UserOrg       `json:"users,omitempty"`
}

func (o *Org) BeforeCreate(tx *gorm.DB) error {
	o.ID = domains.NewOrgID()
	return nil
}
