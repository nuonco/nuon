package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type AppInstaller struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index"`

	OrgID string `json:"org_id" gorm:"notnull"`
	AppID string `json:"app_id" gorm:"notnull"`
	App   App

	Slug string `json:"slug" gorm:"notnull;unique"`

	Metadata AppInstallerMetadata `json:"app_installer_metadata" gorm:"constraint:OnDelete:CASCADE;"`
}

func (a *AppInstaller) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppID()
	}

	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	return nil
}
