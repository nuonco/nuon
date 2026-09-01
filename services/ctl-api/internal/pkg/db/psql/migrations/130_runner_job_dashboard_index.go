package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration130RunnerJobDashboardIndex(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_runner_jobs_dashboard_activity
	ON runner_jobs (org_id, runner_id, created_at DESC, "group")
	WHERE deleted_at = 0`).Error
}
