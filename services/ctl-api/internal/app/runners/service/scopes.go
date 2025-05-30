package service

import "gorm.io/gorm"

func runnerJobPreload(db *gorm.DB) *gorm.DB {
	return db.Preload("RunnerJob", func(db *gorm.DB) *gorm.DB {
		return db.Select(
			"ID",
			"CreatedByID",
			"CreatedAt",
			"UpdatedAt",
			"OrgID",
			"RunnerID",
			"OwnerID",
			"OwnerType",
			"LogStreamID",
			"Status",
			"Type",
			"Group",
			"Operation",
			"Metadata",
		)
	})
}
