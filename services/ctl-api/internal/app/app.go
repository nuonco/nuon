package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type App struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"notnull"`
	CreatedBy   UserToken             `json:"created_by" gorm:"references:Subject"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_app_name,unique"`

	Name        string `json:"name" gorm:"index:idx_app_name,unique"`
	Description string `json:"description"`
	DisplayName string `json:"display_name"`

	OrgID string `json:"org_id" gorm:"index:idx_app_name,unique"`
	Org   Org    `faker:"-" json:"-"`

	NotificationsConfig NotificationsConfig `gorm:"polymorphic:Owner;constraint:OnDelete:CASCADE;" json:"notifications_config,omitempty"`

	Components        []Component        `faker:"components" json:"-" swaggerignore:"true" gorm:"constraint:OnDelete:CASCADE;"`
	Installs          []Install          `faker:"-" json:"-" swaggerignore:"true" gorm:"constraint:OnDelete:CASCADE;"`
	AppInputConfigs   []AppInputConfig   `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	AppSandboxConfigs []AppSandboxConfig `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	AppRunnerConfigs  []AppRunnerConfig  `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	AppConfigs        []AppConfig        `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	AppSecrets        []AppSecret        `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	InstallerApps     []InstallerApp     `json:"-" gorm:"constraint:OnDelete:CASCADE;"`

	Status            string `json:"status"`
	StatusDescription string `json:"status_description"`

	// fields set via after query
	AppInputConfig   AppInputConfig   `json:"input_config" gorm:"-"`
	AppSandboxConfig AppSandboxConfig `json:"sandbox_config" gorm:"-"`
	AppRunnerConfig  AppRunnerConfig  `json:"runner_config" gorm:"-"`

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
	if len(a.AppRunnerConfigs) > 0 {
		a.AppRunnerConfig = a.AppRunnerConfigs[0]
		a.CloudPlatform = a.AppRunnerConfigs[0].CloudPlatform
	}
	if len(a.AppInputConfigs) > 0 {
		a.AppInputConfig = a.AppInputConfigs[0]
	}
	if len(a.AppSandboxConfigs) > 0 {
		a.AppSandboxConfig = a.AppSandboxConfigs[0]
	}

	return nil
}
