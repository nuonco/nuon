package migrations

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (m *Migrations) Migration135BackfillInstallExpectedActualState(ctx context.Context, db *gorm.DB) error {
	statements := []struct {
		name string
		sql  string
	}{
		{
			name: "installs actual app config",
			sql: `
		UPDATE installs AS i
		SET actual_app_config_id = latest.new_app_config_id,
		    actual_app_config_applied_at = latest.finished_at,
		    actual_app_config_workflow_id = latest.id
		FROM (
			SELECT DISTINCT ON (w.owner_id)
				w.id, w.owner_id, w.finished_at, w.status->>'status' AS status,
				w.metadata ->> 'new_app_config_id' AS new_app_config_id
			FROM install_workflows w
			WHERE w.owner_type = 'installs'
			  AND w.type = 'app_branch_config_update'
			  AND w.plan_only = false
			  AND w.deleted_at = 0
			ORDER BY w.owner_id, w.created_at DESC
		) AS latest
		WHERE latest.owner_id = i.id
		  AND i.deleted_at = 0
		  AND i.actual_app_config_id IS NULL
		  AND latest.status = 'success'
		  AND latest.finished_at IS NOT NULL
		  AND latest.new_app_config_id = i.app_config_id
		  AND NOT EXISTS (
			SELECT 1 FROM install_workflow_steps s
			WHERE s.install_workflow_id = latest.id
			  AND s.deleted_at = 0
			  AND (
				s.status->>'status' IN ('error', 'user-skipped', 'cancelled', 'not-attempted', 'approval-denied', 'approval-expired')
				OR (s.status->>'status' = 'discarded' AND s.retried = false)
			  )
		  );`,
		},
		{
			name: "install components expected app config",
			sql: `
		UPDATE install_components AS ic
		SET expected_app_config_id = i.app_config_id
		FROM installs i
		JOIN app_configs ac ON ac.id = i.app_config_id AND ac.deleted_at = 0
		WHERE ic.install_id = i.id
		  AND ic.deleted_at = 0
		  AND i.deleted_at = 0
		  AND ic.expected_app_config_id IS NULL
		  AND ic.component_id = ANY(ac.component_ids);`,
		},
		{
			name: "install components actual deploy",
			sql: `
		UPDATE install_components AS ic
		SET actual_install_deploy_id = latest.id,
		    actual_component_build_id = latest.component_build_id,
		    actual_applied_at = latest.applied_at
		FROM (
			SELECT DISTINCT ON (d.install_component_id)
				d.id, d.install_component_id, d.component_build_id, d.applied_at, d.type
			FROM install_deploys d
			WHERE d.applied_at IS NOT NULL
			  AND d.deleted_at = 0
			  AND d.type <> 'recover'
			ORDER BY d.install_component_id, d.applied_at DESC
		) AS latest
		WHERE latest.install_component_id = ic.id
		  AND ic.deleted_at = 0
		  AND ic.actual_install_deploy_id IS NULL
		  AND latest.type <> 'teardown';`,
		},
		{
			name: "install sandboxes expected sandbox config",
			sql: `
		UPDATE install_sandboxes AS isb
		SET expected_app_sandbox_config_id = sc.id
		FROM installs i
		JOIN (
			SELECT app_config_id, MIN(id) AS id
			FROM app_sandbox_configs
			WHERE deleted_at = 0
			GROUP BY app_config_id
			HAVING COUNT(*) = 1
		) AS sc ON sc.app_config_id = i.app_config_id
		WHERE isb.install_id = i.id
		  AND isb.deleted_at = 0
		  AND i.deleted_at = 0
		  AND isb.expected_app_sandbox_config_id IS NULL;`,
		},
		{
			name: "install sandboxes actual sandbox run",
			sql: `
		UPDATE install_sandboxes AS isb
		SET actual_app_sandbox_config_id = latest.app_sandbox_config_id,
		    actual_install_sandbox_run_id = latest.id,
		    actual_applied_at = latest.applied_at
		FROM (
			SELECT DISTINCT ON (r.install_sandbox_id)
				r.id, r.install_sandbox_id, r.app_sandbox_config_id, r.applied_at, r.run_type
			FROM install_sandbox_runs r
			WHERE r.applied_at IS NOT NULL
			  AND r.install_sandbox_id IS NOT NULL
			  AND r.deleted_at = 0
			ORDER BY r.install_sandbox_id, r.applied_at DESC
		) AS latest
		WHERE latest.install_sandbox_id = isb.id
		  AND isb.deleted_at = 0
		  AND isb.actual_install_sandbox_run_id IS NULL
		  AND latest.run_type IN ('provision', 'reprovision');`,
		},
		{
			name: "install stacks expected app config",
			sql: `
		UPDATE install_stacks AS st
		SET expected_app_config_id = i.app_config_id
		FROM installs i
		WHERE st.install_id = i.id
		  AND st.deleted_at = 0
		  AND i.deleted_at = 0
		  AND st.expected_app_config_id IS NULL
		  AND i.app_config_id IS NOT NULL
		  AND i.app_config_id <> '';`,
		},
		{
			name: "install action workflows expected config",
			sql: `
		UPDATE install_action_workflows AS iaw
		SET expected_action_workflow_config_id = awc.id
		FROM installs i
		JOIN action_workflow_configs awc ON awc.app_config_id = i.app_config_id AND awc.deleted_at = 0
		WHERE iaw.install_id = i.id
		  AND iaw.action_workflow_id = awc.action_workflow_id
		  AND iaw.deleted_at = 0
		  AND i.deleted_at = 0
		  AND iaw.expected_action_workflow_config_id IS NULL;`,
		},
		{
			name: "install action workflows actual config",
			sql: `
		UPDATE install_action_workflows AS iaw
		SET actual_action_workflow_config_id = iaw.expected_action_workflow_config_id,
		    actual_applied_at = iaw.updated_at
		WHERE iaw.deleted_at = 0
		  AND iaw.actual_action_workflow_config_id IS NULL
		  AND iaw.expected_action_workflow_config_id IS NOT NULL;`,
		},
	}

	for _, stmt := range statements {
		res := db.WithContext(ctx).Exec(stmt.sql)
		if res.Error != nil {
			return fmt.Errorf("unable to backfill %s: %w", stmt.name, res.Error)
		}
		m.l.Info("backfilled install expected/actual state",
			zap.String("target", stmt.name),
			zap.Int64("rows", res.RowsAffected))
	}
	return nil
}
