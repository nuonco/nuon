package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration137DropAWSAccountConnections(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`
ALTER TABLE aws_accounts DROP COLUMN IF EXISTS aws_account_connection_id;
DROP TABLE IF EXISTS aws_account_connections;
`).Error
}
