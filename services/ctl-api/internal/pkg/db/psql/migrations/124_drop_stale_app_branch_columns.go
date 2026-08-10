package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Both columns were declared NOT NULL with no default and later removed from the
// AppBranch model. AutoMigrate never drops columns, so long-lived databases still
// reject every insert into app_branches.
func (m *Migrations) Migration124DropStaleAppBranchColumns(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`
		ALTER TABLE app_branches DROP COLUMN IF EXISTS vcs_connection_branch_id;
		ALTER TABLE app_branches DROP COLUMN IF EXISTS connected_github_vcs_config_id;
	`).Error
}
