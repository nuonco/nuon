package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration143DropCloudConnectionLegacyColumns(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`
		ALTER TABLE cloud_connections DROP COLUMN IF EXISTS tenant_id;
		ALTER TABLE cloud_connections DROP COLUMN IF EXISTS identity_provider;
	`).Error
}
