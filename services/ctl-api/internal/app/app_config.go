package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
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
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id" gorm:"notnull;default null"`
	Org   Org    `faker:"-" json:"-"`
	AppID string `json:"app_id"`

	Status            AppConfigStatus `json:"status"`
	StatusDescription string          `json:"status_description" gorm:"notnull;default null"`

	State  string `json:"state"`
	Readme string `json:"readme"`

	// fields that are filled in via after query or views
	Version int `json:"version" gorm:"->;-:migration"`
}

func (a AppConfig) UseView() bool {
	return true
}

func (a AppConfig) ViewVersion() string {
	return "v2"
}

func (i *AppConfig) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name: views.DefaultViewName(db, &AppConfig{}, 2),
			SQL:  viewsql.AppConfigViewV2,
		},
	}
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
