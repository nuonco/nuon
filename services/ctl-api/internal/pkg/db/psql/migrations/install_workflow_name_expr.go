package migrations

// InstallWorkflowNameExpr is the SQL expression body for the
// install_workflows.name STORED generated column — the human-readable
// workflow title rendered by the dashboard and CLI (e.g.
// "Deploying to install (rds_cluster_temporal)", "Runbook run
// (deploy_control_plane)").
//
// It lives in one Go const so future "add a workflow type" PRs touch one
// place: update the const, then add a new migration that drops and re-adds
// the column using this same const value. The model field is declared
// `gorm:"-:migration;->"` so AutoMigrate ignores it (see
// app.Workflow.Name).
//
// Postgres recomputes the column on every INSERT / UPDATE that touches
// `type`, `finished_at`, or `metadata`, so the value tracks lifecycle
// transitions (e.g. "Provisioning install" -> "Provisioned install" when
// finished_at is set) without any application code.
const InstallWorkflowNameExpr = `
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
`

// RebuildInstallWorkflowsNameColumnSQL returns the DDL that drops and re-adds
// the install_workflows.name STORED generated column using
// InstallWorkflowNameExpr. Migration 108 (the initial creation) and any
// future migration that updates the title expression both run this — the
// DROP IF EXISTS keeps it idempotent.
func RebuildInstallWorkflowsNameColumnSQL() string {
	return `
		ALTER TABLE install_workflows DROP COLUMN IF EXISTS name;
		ALTER TABLE install_workflows
		ADD COLUMN name TEXT GENERATED ALWAYS AS (` + InstallWorkflowNameExpr + `) STORED;
	`
}
