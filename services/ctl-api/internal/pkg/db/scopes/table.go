package scopes

import (
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

func WithOverrideTable(name string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.InstanceSet(views.DisableViewsKey, true).Table(name)
	}
}

func WithDisableViews(db *gorm.DB) *gorm.DB {
	return db.InstanceSet(views.DisableViewsKey, true)
}
