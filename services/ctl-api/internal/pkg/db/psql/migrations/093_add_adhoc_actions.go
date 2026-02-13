package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration093AddAdhocActions(ctx context.Context, db *gorm.DB) error {
	alterInstallActionWorkflowRuns := `
		ALTER TABLE install_action_workflow_runs
			ALTER COLUMN install_action_workflow_id DROP NOT NULL,
			ALTER COLUMN action_workflow_config_id DROP NOT NULL;
	`
	if res := db.WithContext(ctx).Exec(alterInstallActionWorkflowRuns); res.Error != nil {
		return res.Error
	}

	alterInstallActionWorkflowRunSteps := `
		ALTER TABLE install_action_workflow_run_steps
			ALTER COLUMN step_id DROP NOT NULL,
			ADD COLUMN IF NOT EXISTS adhoc_config JSONB;
	`
	if res := db.WithContext(ctx).Exec(alterInstallActionWorkflowRunSteps); res.Error != nil {
		return res.Error
	}

	return nil
}
