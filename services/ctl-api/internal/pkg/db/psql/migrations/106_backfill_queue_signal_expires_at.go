package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Migration106BackfillQueueSignalExpiresAt sets expires_at = created_at + 1 hour
// for all existing process_healthcheck and healthcheck signals that don't already
// have an expiry, and sets signal_expires_in on existing emitters so newly emitted
// signals also get an expiry. This prevents 600k+ stale handler workflows from accumulating.
func (m *Migrations) Migration106BackfillQueueSignalExpiresAt(ctx context.Context, db *gorm.DB) error {
	// Backfill existing signals
	if res := db.WithContext(ctx).Exec(`
		UPDATE queue_signals
		SET expires_at = created_at + INTERVAL '1 hour'
		WHERE type IN ('process_healthcheck', 'healthcheck')
		  AND expires_at IS NULL
		  AND deleted_at = 0;
	`); res.Error != nil {
		return res.Error
	}

	// Backfill existing emitters so future emissions include an expiry
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
