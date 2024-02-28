package app

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/config"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type AppInstaller struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   UserToken             `json:"created_by" gorm:"references:Subject"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_app_installer_slug,unique" json:"-"`

	OrgID string `json:"org_id" gorm:"notnull"`
	Org   Org    `faker:"-" json:"-"`
	AppID string `json:"app_id" gorm:"notnull"`
	App   App

	Slug string `json:"slug" gorm:"index:idx_app_installer_slug,unique"`

	Metadata AppInstallerMetadata `json:"app_installer_metadata" gorm:"constraint:OnDelete:CASCADE;"`

	// filled in via after query
	InstallerURL string `json:"installer_url" gorm:"-"`
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

func (a *AppInstaller) AfterQuery(tx *gorm.DB) error {
	cfg, err := config.FromContext(tx.Statement.Context)
	if err != nil {
		return nil
	}

	a.InstallerURL = fmt.Sprintf("%s/installer/%s", cfg.InstallerBaseURL, a.Slug)
	return nil
}
