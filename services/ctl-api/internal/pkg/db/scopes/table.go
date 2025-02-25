package scopes

import (
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
)

func WithOverrideTable(name string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Table(name)
	}
}

// DisableViews is a scope that removes usage of views. Usage for when the view returns too much data or is slow.
func WithDisableViews(db *gorm.DB) *gorm.DB {
	return db.Set(views.DisableViewsKey, true)
}
