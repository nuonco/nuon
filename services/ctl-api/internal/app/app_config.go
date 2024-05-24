package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type AppConfigStatus string

const (
	AppConfigStatusActive   AppConfigStatus = "active"
	AppConfigStatusPending  AppConfigStatus = "pending"
	AppConfigStatusSyncing  AppConfigStatus = "syncing"
	AppConfigStatusError    AppConfigStatus = "error"
	AppConfigStatusOutdated AppConfigStatus = "outdated"
)

type AppConfigFmt string

const (
	AppConfigFmtToml AppConfigFmt = "toml"
)

type AppConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   UserToken             `json:"created_by" gorm:"references:Subject"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id" gorm:"notnull;default null"`
	Org   Org    `faker:"-" json:"-"`
	AppID string `json:"app_id"`

	Format  AppConfigFmt `json:"format" gorm:"notnull;default null"`
	Content string       `json:"content" gorm:"notnull;default null"`

	Status            AppConfigStatus `json:"status" gorm:"notnull;default null"`
	StatusDescription string          `json:"status_description" gorm:"notnull;default null"`

	GeneratedTerraform string `json:"generated_terraform"`

	// fields that are filled in via after query or views
	Version int `json:"version" gorm:"->;-:migration"`
}

func (a AppConfig) ViewName() string {
	return "app_configs_view"
}

func (a *AppConfig) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
