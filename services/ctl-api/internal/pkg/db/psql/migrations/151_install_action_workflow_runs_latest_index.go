package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration151InstallActionWorkflowRunsLatestIndex(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_install_action_workflow_runs_latest
	ON install_action_workflow_runs (install_action_workflow_id, created_at DESC)
	WHERE deleted_at = 0`).Error
}
