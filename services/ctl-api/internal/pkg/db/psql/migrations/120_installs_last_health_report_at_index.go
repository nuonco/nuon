package migrations

import (
	"context"

	"gorm.io/gorm"
)

// The component health sweep runs on a cron and selects installs by
// last_health_report_at alone, which seq-scanned the whole installs table every
// run. Partial predicate mirrors the sweep's own filters exactly — a mismatch
// here silently leaves the index unused.
func (m *Migrations) Migration120InstallsLastHealthReportAtIndex(ctx context.Context, db *gorm.DB) error {
	res := db.WithContext(ctx).Exec(
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_installs_last_health_report_at
ON installs (last_health_report_at)
WHERE deleted_at = 0
  AND last_health_report_at IS NOT NULL`,
	)
	return res.Error
}
