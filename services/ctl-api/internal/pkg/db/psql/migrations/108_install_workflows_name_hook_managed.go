package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Migration108InstallWorkflowsNameHookManaged converts install_workflows.name
// from a STORED generated column (previous, never-shipped migration) to a
// plain TEXT column populated by Workflow.BeforeSave / computeWorkflowName.
//
// The backfill below is the last time this title expression appears in SQL.
// Going forward the source of truth is computeWorkflowName (Go); new rows
// get their name from the hook on INSERT/UPDATE.
func (m *Migrations) Migration108InstallWorkflowsNameHookManaged(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Exec(`
		ALTER TABLE install_workflows DROP COLUMN IF EXISTS name;
		ALTER TABLE install_workflows ADD COLUMN name TEXT;

		UPDATE install_workflows SET name = (
			CASE
			WHEN type = 'action_workflow_run' AND COALESCE(metadata->'adhoc_action', '') <> ''
				THEN 'Adhoc action run (' || COALESCE(metadata->'install_action_workflow_name', '') || ')'
			ELSE
				COALESCE(
					CASE
					WHEN finished_at IS NULL THEN
						CASE type
							WHEN 'provision' THEN 'Provisioning install'
							WHEN 'reprovision' THEN 'Reprovisioning install'
							WHEN 'deprovision' THEN 'Deprovisioning install'
							WHEN 'manual_deploy' THEN 'Deploying to install'
							WHEN 'drift_run' THEN 'Deploying to install'
							WHEN 'input_update' THEN 'Input Update'
							WHEN 'teardown_components' THEN 'Tearing down all components'
							WHEN 'deploy_components' THEN 'Deploying all components'
							WHEN 'reprovision_sandbox' THEN 'Reprovisioning sandbox'
							WHEN 'drift_run_reprovision_sandbox' THEN 'Reprovisioning sandbox'
							WHEN 'sync_secrets' THEN 'Syncing secrets'
							WHEN 'action_workflow_run' THEN 'Action run'
							WHEN 'app_config_build' THEN 'Building app config components'
							WHEN 'runbook_run' THEN 'Running runbook'
						END
					ELSE
						CASE type
							WHEN 'provision' THEN 'Provisioned install'
							WHEN 'reprovision' THEN 'Reprovisioned install'
							WHEN 'reprovision_sandbox' THEN 'Reprovisioned sandbox'
							WHEN 'drift_run_reprovision_sandbox' THEN 'Reprovisioned sandbox'
							WHEN 'deprovision' THEN 'Deprovisioned install'
							WHEN 'manual_deploy' THEN 'Deployed to install'
							WHEN 'drift_run' THEN 'Deployed to install'
							WHEN 'input_update' THEN 'Updated Input'
							WHEN 'teardown_components' THEN 'Tore down all components'
							WHEN 'deploy_components' THEN 'Deployed all components'
							WHEN 'sync_secrets' THEN 'Synced secrets'
							WHEN 'action_workflow_run' THEN 'Action run'
							WHEN 'app_config_build' THEN 'Built app config components'
							WHEN 'runbook_run' THEN 'Runbook run'
						END
					END,
					REPLACE(type, '_', ' ')
				)
				|| COALESCE(' (' || NULLIF(metadata->'workflow-name-suffix', '') || ')', '')
				|| CASE WHEN type = 'action_workflow_run'
						THEN COALESCE(' (' || NULLIF(metadata->'install_action_workflow_name', '') || ')', '')
						ELSE '' END
				|| CASE WHEN type = 'runbook_run'
						THEN COALESCE(' (' || NULLIF(metadata->'runbook_name', '') || ')', '')
						ELSE '' END
			END
		);

		CREATE INDEX IF NOT EXISTS idx_install_workflows_name ON install_workflows (name);
	`).Error
}
