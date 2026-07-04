package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration117AppBranchUniqueWithDeletedAt(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`
		DROP INDEX IF EXISTS idx_app_branch_name_per_app;
		CREATE UNIQUE INDEX idx_app_branch_name_per_app ON app_branches (app_id, name, deleted_at);
	`).Error
}
