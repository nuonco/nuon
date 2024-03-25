package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type App struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"notnull"`
	CreatedBy   UserToken             `json:"created_by" gorm:"references:Subject"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_app_name,unique"`

	Name string `json:"name" gorm:"index:idx_app_name,unique"`

	OrgID string `json:"org_id" gorm:"index:idx_app_name,unique"`
	Org   Org    `faker:"-" json:"-"`

	Components        []Component        `faker:"-" json:"-" swaggerignore:"true" gorm:"constraint:OnDelete:CASCADE;"`
	Installs          []Install          `faker:"-" json:"-" swaggerignore:"true" gorm:"constraint:OnDelete:CASCADE;"`
	Installers        []AppInstaller     `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	AppInputConfigs   []AppInputConfig   `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	AppSandboxConfigs []AppSandboxConfig `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	AppRunnerConfigs  []AppRunnerConfig  `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	AppConfigs        []AppConfig        `json:"-" gorm:"constraint:OnDelete:CASCADE;"`

	Status            string `json:"status"`
	StatusDescription string `json:"status_description"`

	// filled in via after query
	CloudPlatform CloudPlatform `json:"cloud_platform" gorm:"-"`
}

func (a *App) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (a *App) AfterQuery(tx *gorm.DB) error {
	a.CloudPlatform = CloudPlatformUnknown
	if len(a.AppRunnerConfigs) < 1 {
		return nil
	}

	a.CloudPlatform = a.AppRunnerConfigs[0].CloudPlatform
	return nil
}
