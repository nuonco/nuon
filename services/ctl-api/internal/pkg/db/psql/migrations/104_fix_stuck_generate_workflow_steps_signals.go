package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration104FixStuckGenerateWorkflowStepsSignals(ctx context.Context, db *gorm.DB) error {
	res := db.WithContext(ctx).Exec(`
		UPDATE queue_signals
		SET status = jsonb_set(status, '{status}', '"success"')
		WHERE type = 'generate-workflow-steps'
		AND (status->>'status') NOT IN ('success', 'error', 'cancelled')
		AND created_at < NOW() - INTERVAL '15 minute';
	`)
	return res.Error
}
