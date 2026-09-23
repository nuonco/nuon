package migrations

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (m *Migrations) Migration136MigrateAppConfigRef(ctx context.Context, db *gorm.DB) error {
	backfills := []struct {
		name string
		sql  string
	}{
		{
			name: "installs app_config_ref from actual columns",
			sql: `
UPDATE installs
SET app_config_ref = jsonb_strip_nulls(jsonb_build_object(
    'expected_config_id', NULLIF(app_config_id, ''),
    'applied_config_id',  actual_app_config_id,
    'applied_config_at',  actual_app_config_applied_at,
    'applied_config_by_type', CASE WHEN actual_app_config_workflow_id IS NOT NULL THEN 'install_workflows' ELSE NULL END,
    'applied_config_by_id',   actual_app_config_workflow_id
))
WHERE deleted_at = 0
  AND (app_config_ref IS NULL OR app_config_ref = 'null'::jsonb OR app_config_ref = '{}'::jsonb);`,
		},
		{
			name: "install_stacks app_config_ref from expected/actual columns",
			sql: `
UPDATE install_stacks
SET app_config_ref = jsonb_strip_nulls(jsonb_build_object(
    'expected_config_id', expected_app_config_id,
    'applied_config_id',  actual_app_config_id,
    'applied_config_at',  actual_applied_at,
    'applied_config_by_type', CASE WHEN actual_install_stack_version_id IS NOT NULL THEN 'install_stack_versions' ELSE NULL END,
    'applied_config_by_id',   actual_install_stack_version_id
))
WHERE deleted_at = 0
  AND (app_config_ref IS NULL OR app_config_ref = 'null'::jsonb OR app_config_ref = '{}'::jsonb);`,
		},
		{
			name: "install_sandboxes app_config_ref from expected/actual columns (normalised to app_config_id)",
			sql: `
UPDATE install_sandboxes isb
SET app_config_ref = jsonb_strip_nulls(jsonb_build_object(
    'expected_config_id', exp_ac.app_config_id,
    'applied_config_id',  act_ac.app_config_id,
    'applied_config_at',  isb.actual_applied_at,
    'applied_config_by_type', CASE WHEN isb.actual_install_sandbox_run_id IS NOT NULL THEN 'install_sandbox_runs' ELSE NULL END,
    'applied_config_by_id',   isb.actual_install_sandbox_run_id
))
FROM app_sandbox_configs exp_ac
LEFT JOIN app_sandbox_configs act_ac ON act_ac.id = isb.actual_app_sandbox_config_id AND act_ac.deleted_at = 0
WHERE exp_ac.id = isb.expected_app_sandbox_config_id
  AND exp_ac.deleted_at = 0
  AND isb.deleted_at = 0
  AND (isb.app_config_ref IS NULL OR isb.app_config_ref = 'null'::jsonb OR isb.app_config_ref = '{}'::jsonb);`,
		},
		{
			name: "install_sandboxes app_config_ref (no expected sandbox config – use install app_config_id)",
			sql: `
UPDATE install_sandboxes isb
SET app_config_ref = jsonb_strip_nulls(jsonb_build_object(
    'expected_config_id', i.app_config_id,
    'applied_config_at',  isb.actual_applied_at,
    'applied_config_by_type', CASE WHEN isb.actual_install_sandbox_run_id IS NOT NULL THEN 'install_sandbox_runs' ELSE NULL END,
    'applied_config_by_id',   isb.actual_install_sandbox_run_id
))
FROM installs i
WHERE i.id = isb.install_id
  AND isb.expected_app_sandbox_config_id IS NULL
  AND isb.deleted_at = 0
  AND i.deleted_at = 0
  AND (isb.app_config_ref IS NULL OR isb.app_config_ref = 'null'::jsonb OR isb.app_config_ref = '{}'::jsonb);`,
		},
		{
			name: "install_components app_config_ref from expected/actual columns",
			sql: `
UPDATE install_components
SET app_config_ref = jsonb_strip_nulls(jsonb_build_object(
    'expected_config_id', expected_app_config_id,
    'applied_config_id',  expected_app_config_id,
    'applied_config_at',  actual_applied_at,
    'applied_config_by_type', CASE WHEN actual_install_deploy_id IS NOT NULL THEN 'install_deploys' ELSE NULL END,
    'applied_config_by_id',   actual_install_deploy_id
))
WHERE deleted_at = 0
  AND (app_config_ref IS NULL OR app_config_ref = 'null'::jsonb OR app_config_ref = '{}'::jsonb);`,
		},
		{
			name: "install_action_workflows app_config_ref (normalised to app_config_id via action_workflow_config)",
			sql: `
UPDATE install_action_workflows iaw
SET app_config_ref = jsonb_strip_nulls(jsonb_build_object(
    'expected_config_id', exp_awc.app_config_id,
    'applied_config_id',  act_awc.app_config_id,
    'applied_config_at',  iaw.actual_applied_at,
    'applied_config_by_type', CASE WHEN iaw.actual_action_workflow_config_id IS NOT NULL THEN 'install_action_workflow_runs' ELSE NULL END,
    'applied_config_by_id',   iaw.actual_action_workflow_config_id
))
FROM action_workflow_configs exp_awc
LEFT JOIN action_workflow_configs act_awc ON act_awc.id = iaw.actual_action_workflow_config_id AND act_awc.deleted_at = 0
WHERE exp_awc.id = iaw.expected_action_workflow_config_id
  AND exp_awc.deleted_at = 0
  AND iaw.deleted_at = 0
  AND (iaw.app_config_ref IS NULL OR iaw.app_config_ref = 'null'::jsonb OR iaw.app_config_ref = '{}'::jsonb);`,
		},
		{
			name: "install_action_workflows app_config_ref (no expected config – use install app_config_id)",
			sql: `
UPDATE install_action_workflows iaw
SET app_config_ref = jsonb_strip_nulls(jsonb_build_object(
    'expected_config_id', i.app_config_id
))
FROM installs i
WHERE i.id = iaw.install_id
  AND iaw.expected_action_workflow_config_id IS NULL
  AND iaw.deleted_at = 0
  AND i.deleted_at = 0
  AND (iaw.app_config_ref IS NULL OR iaw.app_config_ref = 'null'::jsonb OR iaw.app_config_ref = '{}'::jsonb);`,
		},
	}

	for _, b := range backfills {
		res := db.WithContext(ctx).Exec(b.sql)
		if res.Error != nil {
			return fmt.Errorf("unable to backfill %s: %w", b.name, res.Error)
		}
		m.l.Info("backfilled app_config_ref",
			zap.String("target", b.name),
			zap.Int64("rows", res.RowsAffected))
	}

	drops := []struct {
		name string
		sql  string
	}{
		{
			name: "drop installs old actual columns",
			sql: `
ALTER TABLE installs
    DROP COLUMN IF EXISTS actual_app_config_id,
    DROP COLUMN IF EXISTS actual_app_config_applied_at,
    DROP COLUMN IF EXISTS actual_app_config_workflow_id;`,
		},
		{
			name: "drop install_stacks old expected/actual columns",
			sql: `
ALTER TABLE install_stacks
    DROP COLUMN IF EXISTS expected_app_config_id,
    DROP COLUMN IF EXISTS actual_app_config_id,
    DROP COLUMN IF EXISTS actual_install_stack_version_id,
    DROP COLUMN IF EXISTS actual_applied_at;`,
		},
		{
			name: "drop install_sandboxes old expected/actual columns",
			sql: `
ALTER TABLE install_sandboxes
    DROP COLUMN IF EXISTS expected_app_sandbox_config_id,
    DROP COLUMN IF EXISTS actual_app_sandbox_config_id,
    DROP COLUMN IF EXISTS actual_install_sandbox_run_id,
    DROP COLUMN IF EXISTS actual_applied_at;`,
		},
		{
			name: "drop install_components old expected/actual columns",
			sql: `
ALTER TABLE install_components
    DROP COLUMN IF EXISTS expected_app_config_id,
    DROP COLUMN IF EXISTS actual_install_deploy_id,
    DROP COLUMN IF EXISTS actual_component_build_id,
    DROP COLUMN IF EXISTS actual_applied_at;`,
		},
		{
			name: "drop install_action_workflows old expected/actual columns",
			sql: `
ALTER TABLE install_action_workflows
    DROP COLUMN IF EXISTS expected_action_workflow_config_id,
    DROP COLUMN IF EXISTS actual_action_workflow_config_id,
    DROP COLUMN IF EXISTS actual_applied_at;`,
		},
	}

	for _, d := range drops {
		if err := db.WithContext(ctx).Exec(d.sql).Error; err != nil {
			return fmt.Errorf("unable to %s: %w", d.name, err)
		}
		m.l.Info("dropped old columns", zap.String("target", d.name))
	}

	return nil
}
