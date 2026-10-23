package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration148DropDeviceCodes(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`DROP TABLE IF EXISTS device_codes`).Error
}
