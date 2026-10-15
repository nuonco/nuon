package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration143ActivityFeedIndexes(ctx context.Context, db *gorm.DB) error {
	statements := []string{
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_install_runbook_runs_install_workflow_id
ON install_runbook_runs (install_workflow_id)
WHERE deleted_at = 0`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_install_action_workflow_runs_activity_non_cron
ON install_action_workflow_runs (org_id, install_id, created_at DESC, id DESC)
WHERE deleted_at = 0 AND trigger_type <> 'cron'`,
	}

	for _, statement := range statements {
		if err := db.WithContext(ctx).Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}
