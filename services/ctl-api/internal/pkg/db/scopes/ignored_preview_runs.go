package scopes

import (
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func ignoredPreviewRun() app.AppBranchRun {
	return app.AppBranchRun{
		RunType: app.AppBranchRunTypeGitPreview,
		Status:  string(app.StatusNotAttempted),
	}
}

func ExcludeIgnoredPreviewRuns(db *gorm.DB) *gorm.DB {
	return db.Not(ignoredPreviewRun())
}

func ExcludeIgnoredPreviewWorkflows(appBranchID string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		ignored := db.Session(&gorm.Session{NewDB: true}).
			Model(&app.AppBranchRun{}).
			Select("workflow_id").
			Where(app.AppBranchRun{AppBranchID: appBranchID}).
			Where(ignoredPreviewRun()).
			Where("workflow_id IS NOT NULL")
		return db.Where("id NOT IN (?)", ignored)
	}
}
