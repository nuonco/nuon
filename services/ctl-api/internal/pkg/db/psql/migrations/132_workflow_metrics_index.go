package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration132WorkflowMetricsIndex(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_install_workflows_metrics_current
	ON install_workflows (type, owner_id, created_at)
	WHERE owner_type = 'installs' AND deleted_at = 0 AND plan_only IS NOT TRUE
	  AND (status->>'status' IN ('pending', 'queued', 'in-progress', 'retrying', 'approval-awaiting', 'failed-pending-retry')
	    OR (status->>'status' = 'error' AND status->'metadata'->>'awaiting_retry' = 'true'))`).Error
}
