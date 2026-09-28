package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration144RetireCloudConnectionCronEmitters(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`
		UPDATE queue_emitters
		SET deleted_at = EXTRACT(EPOCH FROM now())::bigint, updated_at = now()
		WHERE deleted_at = 0
			AND mode = 'cron'
			AND signal_type = 'cloud_connection_reverify'
	`).Error
}
