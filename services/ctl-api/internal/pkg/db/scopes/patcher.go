package scopes

import (
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/patcher"
)

func WithPatcher(exclusions []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if exclusions == nil || len(exclusions) == 0 {
			db = db.Set(patcher.PatcherExclusionsKey, []string{})
		}

		if len(exclusions) > 0 {
			db = db.Set(patcher.PatcherExclusionsKey, exclusions)
		}

		db = db.Set(patcher.PatcherEnabledKey, true)
		return db
	}
}
