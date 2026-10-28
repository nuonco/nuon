package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration151RunnerJobLogStreamIndex(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_runner_jobs_log_stream_id
	ON runner_jobs (log_stream_id)
	WHERE deleted_at = 0`).Error
}
