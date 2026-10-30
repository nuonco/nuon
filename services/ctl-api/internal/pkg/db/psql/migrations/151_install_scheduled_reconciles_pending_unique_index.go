package migrations

import (
	"context"

	"gorm.io/gorm"
)

// At most one pending scheduled reconcile per install. The service layer also
// checks this before creating a row, but the partial unique index is the
// source of truth under concurrent requests.
func (m *Migrations) Migration151InstallScheduledReconcilesPendingUniqueIndex(ctx context.Context, db *gorm.DB) error {
	res := db.WithContext(ctx).Exec(
		`CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_install_scheduled_reconciles_one_pending
ON install_scheduled_reconciles (install_id)
WHERE deleted_at = 0
  AND status = 'pending'`,
	)
	return res.Error
}
