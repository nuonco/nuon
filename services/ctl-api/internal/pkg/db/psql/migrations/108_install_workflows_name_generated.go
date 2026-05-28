package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Migration108InstallWorkflowsNameGenerated adds a STORED generated column
// `name` to install_workflows that contains the human-readable workflow
// title (e.g. "Deploying to install (rds_cluster_temporal)", "Runbook run
// (deploy_control_plane)"). The expression mirrors what app.Workflow used
// to compute in AfterQuery — moving it into the schema lets the workflow
// search filter substring-match the displayed title, and removes the need
// to recompute it in Go on every query.
//
// The column is STORED so Postgres recomputes it automatically whenever
// `finished_at` or `metadata` changes, which is the only way the title
// can shift over a workflow's lifetime (e.g. "Provisioning install" ->
// "Provisioned install" when finished_at is set).
//
// The Workflow model marks Name with `gorm:"-:migration"` so AutoMigrate
// does not also try to create a regular `name` column.
func (m *Migrations) Migration108InstallWorkflowsNameGenerated(ctx context.Context, db *gorm.DB) error {
	// Drop first to make the migration safely re-runnable if the expression
	// is later updated (a follow-up migration drops/re-adds).
	if res := db.WithContext(ctx).Exec(`ALTER TABLE install_workflows DROP COLUMN IF EXISTS name;`); res.Error != nil {
		return res.Error
	}

	return db.WithContext(ctx).Exec(`
		ALTER TABLE install_workflows
		ADD COLUMN name TEXT GENERATED ALWAYS AS (
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
		) STORED;
	`).Error
}
