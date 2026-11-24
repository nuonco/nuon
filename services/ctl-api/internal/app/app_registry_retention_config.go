package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type AppRegistryRetentionConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppID       string `json:"app_id,omitzero" temporaljson:"app_id,omitzero,omitempty"`
	AppConfigID string `json:"app_config_id,omitzero" temporaljson:"app_config_id,omitzero,omitempty"`

	Mode        *string `json:"mode,omitempty" gorm:"default null" temporaljson:"mode,omitempty" extensions:"x-nullable"`
	RetainCount *int    `json:"retain_count,omitempty" gorm:"default null" temporaljson:"retain_count,omitempty" extensions:"x-nullable"`

	GracePeriodDays *int `json:"grace_period_days,omitempty" gorm:"default null" temporaljson:"grace_period_days,omitempty" extensions:"x-nullable"`
}

func (a *AppRegistryRetentionConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppRegistryRetentionConfig{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (a *AppRegistryRetentionConfig) BeforeCreate(tx *gorm.DB) error {
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

func ResolveRegistryGracePeriod(appCfg *AppRegistryRetentionConfig) time.Duration {
	days := config.DefaultRegistryRetentionGracePeriodDays
	if appCfg != nil && appCfg.GracePeriodDays != nil {
		days = *appCfg.GracePeriodDays
	}
	return time.Duration(days) * 24 * time.Hour
}

func ResolveRegistryRetention(installConfig *InstallConfig, appCfg *AppRegistryRetentionConfig, ccc *ComponentConfigConnection) (string, int) {
	mode := config.DefaultRegistryRetentionMode
	retainCount := config.DefaultRegistryRetentionRetainCount

	if appCfg != nil {
		if appCfg.Mode != nil {
			mode = *appCfg.Mode
		}
		if appCfg.RetainCount != nil {
			retainCount = *appCfg.RetainCount
		}
	}
	if installConfig != nil && installConfig.RegistryRetentionMode != nil {
		mode = *installConfig.RegistryRetentionMode
	}
	if ccc != nil && ccc.RegistryRetainCount != nil {
		retainCount = *ccc.RegistryRetainCount
	}

	return mode, retainCount
}
