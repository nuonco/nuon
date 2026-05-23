package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Migration107BackfillEmitterSignalExpiresIn sets signal_expires_in = 1 hour
// for all emitters that don't already have one, so emitted signals expire
// instead of accumulating indefinitely.
func (m *Migrations) Migration107BackfillEmitterSignalExpiresIn(ctx context.Context, db *gorm.DB) error {
	if res := db.WithContext(ctx).Exec(`
		UPDATE queue_emitters
		SET signal_expires_in = 3600000000000
		WHERE (signal_expires_in IS NULL OR signal_expires_in = 0)
		  AND deleted_at = 0;
	`); res.Error != nil {
		return res.Error
	}

	return nil
}
