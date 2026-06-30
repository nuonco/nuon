package app

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/viewsql"
)

type AppLabelKey struct {
	Key         string         `json:"key" gorm:"->;-:migration"`
	Values      pq.StringArray `json:"values" gorm:"->;-:migration;type:text[]"`
	EntityTypes pq.StringArray `json:"entity_types" gorm:"->;-:migration;type:text[]"`
	UsageCount  int            `json:"usage_count" gorm:"->;-:migration"`
	FirstUsedAt time.Time      `json:"first_used_at" gorm:"->;-:migration"`
	AppID       string         `json:"app_id" gorm:"->;-:migration"`
	OrgID       string         `json:"org_id" gorm:"->;-:migration"`
}

func (a *AppLabelKey) UseView() bool {
	return true
}

func (a *AppLabelKey) ViewVersion() string {
	return "v1"
}

func (a *AppLabelKey) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.DefaultViewName(db, &AppLabelKey{}, 1),
			SQL:           viewsql.AppLabelKeysViewV1,
			AlwaysReapply: true,
		},
	}
}

func (a *AppLabelKey) Indexes(db *gorm.DB) []migrations.Index {
	return nil
}

func (a *AppLabelKey) BeforeCreate(tx *gorm.DB) error {
	return nil
}
