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
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_app_installer_slug,unique" json:"-"`

	OrgID string `json:"org_id" gorm:"notnull"`
	AppID string `json:"app_id" gorm:"notnull"`
	App   App

	Slug string `json:"slug" gorm:"index:idx_app_installer_slug,unique"`

	Metadata AppInstallerMetadata `json:"app_installer_metadata" gorm:"constraint:OnDelete:CASCADE;"`
}

func (a *AppInstaller) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppID()
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	return nil
}
