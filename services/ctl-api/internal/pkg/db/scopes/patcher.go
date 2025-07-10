package scopes

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/patcher"
	"gorm.io/gorm"
)

func WithPatcher(options patcher.PatcherOptions) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db = db.InstanceSet(patcher.PatcherOptionsKey, options.Exclusions)
		db = db.InstanceSet(patcher.PatcherEnabledKey, true)
		return db
	}
}
