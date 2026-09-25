package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration138CloudConnectionRequestedCapabilities(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`
ALTER TABLE cloud_connections ADD COLUMN IF NOT EXISTS requested_capabilities jsonb NOT NULL DEFAULT '[]'::jsonb;
UPDATE cloud_connections SET requested_capabilities = capabilities;
`).Error
}
