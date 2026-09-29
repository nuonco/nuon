package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration106BackfillQueueSignalExpiresAt(ctx context.Context, db *gorm.DB) error {
	if res := db.WithContext(ctx).Exec(`
		UPDATE queue_signals
		SET expires_at = created_at + INTERVAL '1 hour'
		WHERE type IN ('process_healthcheck', 'healthcheck')
		  AND expires_at IS NULL
		  AND deleted_at = 0;
	`); res.Error != nil {
		return res.Error
	}

	if res := db.WithContext(ctx).Exec(`
		UPDATE queue_emitters
		SET signal_expires_in = 3600000000000
		WHERE signal_type IN ('process_healthcheck', 'healthcheck')
		  AND (signal_expires_in IS NULL OR signal_expires_in = 0)
		  AND deleted_at = 0;
	`); res.Error != nil {
		return res.Error
	}

	return nil
}
