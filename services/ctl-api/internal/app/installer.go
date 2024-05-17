package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type InstallerType string

const (
	InstallerTypeSelfHosted InstallerType = "self_hosted"
)

type Installer struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   UserToken             `json:"created_by" gorm:"references:Subject"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	OrgID string `json:"org_id" gorm:"notnull"`
	Org   Org    `faker:"-" json:"-"`

	Apps []App `json:"apps" gorm:"many2many:installer_apps;constraint:OnDelete:CASCADE;"`

	Type     InstallerType     `json:"type"`
	Metadata InstallerMetadata `json:"metadata" gorm:"constraint:OnDelete:CASCADE;"`
}

func (a *Installer) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewInstallerID()
	}

	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (a *Installer) AfterQuery(tx *gorm.DB) error {
	return nil
}
