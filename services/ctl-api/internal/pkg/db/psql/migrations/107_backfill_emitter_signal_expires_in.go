package migrations

import (
	"context"

	"gorm.io/gorm"
)

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
