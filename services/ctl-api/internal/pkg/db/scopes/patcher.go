package scopes

import (
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/patcher"
)

func WithPatcher(options patcher.PatcherOptions) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db = db.InstanceSet(patcher.PatcherOptionsKey, options.Exclusions)
		db = db.InstanceSet(patcher.PatcherEnabledKey, true)
		return db
	}
}
