package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration140RetryInstallAppBranchConnections(ctx context.Context, db *gorm.DB) error {
	return backfillInstallAppBranchConnections(ctx, db)
}
