package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration115VcsConnectionUniqueWithDeletedAt(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`
		DROP INDEX IF EXISTS idx_github_install_id;
		CREATE UNIQUE INDEX idx_github_install_id ON vcs_connections (org_id, github_install_id, deleted_at);
	`).Error
}
