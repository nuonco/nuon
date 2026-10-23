package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration147InstallStatesUnarchivedIndex(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_install_states_unarchived
ON install_states (install_id)
WHERE NOT archived AND deleted_at = 0`).Error
}
