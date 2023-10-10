package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type AppInstallerMetadata struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index"`

	AppInstallerID string `json:"app_installer_id" gorm:"notnull"`

	Name             string `json:"name" gorm:"notnull"`
	Description      string `json:"description" gorm:"description"`
	DocumentationURL string `json:"documentation_url" gorm:"notnull"`
	LogoURL          string `json:"logo_url" gorm:"logo_url"`
	GithubURL        string `json:"github_url" gorm:"github_url"`
}

func (a *AppInstallerMetadata) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppID()
	}

	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	return nil
}
