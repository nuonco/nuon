package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration128RunnerJobAvailableTailIndexes(ctx context.Context, db *gorm.DB) error {
	statements := []string{
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_runner_jobs_available_tail
ON runner_jobs (runner_id, created_at DESC)
WHERE deleted_at = 0 AND status = 'available'`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_runner_jobs_available_group_tail
ON runner_jobs (runner_id, "group", created_at DESC)
WHERE deleted_at = 0 AND status = 'available'`,
	}

	for _, statement := range statements {
		if err := db.WithContext(ctx).Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}
