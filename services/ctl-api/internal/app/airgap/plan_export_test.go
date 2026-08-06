package airgap

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

func TestClearTerraformPlanArtifacts(t *testing.T) {
	plan, err := clearTerraformPlanArtifacts(json.RawMessage(`{"future_field":"preserved","deploy_plan":{"apply_plan_contents":"binary-plan","apply_plan_display":"display","terraform":{"plan_json":"cGxhbg==","vars":{"region":"us-west-2"},"future_terraform_field":true}}}`))
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(plan, &decoded))
	require.Equal(t, "preserved", decoded["future_field"])
	deploy := decoded["deploy_plan"].(map[string]any)
	require.Equal(t, "", deploy["apply_plan_contents"])
	require.Equal(t, "", deploy["apply_plan_display"])
	terraformPlan := deploy["terraform"].(map[string]any)
	require.Nil(t, terraformPlan["plan_json"])
	require.Equal(t, true, terraformPlan["future_terraform_field"])
}

func planExportTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.Use(views.NewViewsPlugin([]interface{}{&app.RunnerJob{}})))

	for _, ddl := range []string{
		`CREATE TABLE runner_jobs (
			id text primary key, created_by_id text, created_at datetime, updated_at datetime, deleted_at integer not null default 0,
			org_id text, runner_id text, runner_process_id text, owner_id text, owner_type text, log_stream_id text,
			queue_timeout integer, available_timeout integer, execution_timeout integer, overall_timeout integer,
			max_executions integer, status text, status_description text, status_v2 json,
			type text, "group" text, operation text, executor text,
			started_at datetime, finished_at datetime, metadata text,
			execution_count integer, final_runner_job_execution_id text, outputs json
		)`,
		`CREATE VIEW runner_jobs_view_v2 AS SELECT * FROM runner_jobs`,
		`CREATE TABLE runners (id text primary key, runner_group_id text, deleted_at integer not null default 0)`,
		`CREATE TABLE runner_groups (id text primary key, owner_id text, owner_type text, deleted_at integer not null default 0)`,
		`CREATE TABLE runner_job_plans (
			id text primary key, created_at datetime, deleted_at integer not null default 0,
			runner_job_id text, plan_json text, composite_plan json
		)`,
	} {
		require.NoError(t, db.Exec(ddl).Error)
	}
	return db
}

func seedRunnerJob(t *testing.T, db *gorm.DB, id, runnerID, group, jobType, operation, ownerType, ownerID string, createdAt time.Time, plan string) {
	t.Helper()
	require.NoError(t, db.Exec(
		`INSERT INTO runner_jobs (id, created_at, runner_id, owner_id, owner_type, type, "group", operation) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, createdAt, runnerID, ownerID, ownerType, jobType, group, operation,
	).Error)
	if plan != "" {
		require.NoError(t, db.Exec(
			`INSERT INTO runner_job_plans (id, created_at, runner_job_id, composite_plan) VALUES (?, ?, ?, ?)`,
			"plan-"+id, createdAt, id, plan,
		).Error)
	}
}

func TestExportInstallStepsQueriesRunnerJobView(t *testing.T) {
	db := planExportTestDB(t)
	installID := "inl00000000000000000000001"

	require.NoError(t, db.Exec(`INSERT INTO runner_groups (id, owner_id, owner_type) VALUES ('rg1', ?, 'installs')`, installID).Error)
	require.NoError(t, db.Exec(`INSERT INTO runners (id, runner_group_id) VALUES ('r1', 'rg1')`).Error)

	base := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	sandboxPlan := `{"sandbox_run_plan":{}}`
	seedRunnerJob(t, db, "job00000000000000000000001", "r1", string(app.RunnerJobGroupSandbox),
		"sandbox-terraform", "create-apply-plan", "install_sandbox_runs", "run00000000000000000000001", base, sandboxPlan)
	seedRunnerJob(t, db, "job00000000000000000000002", "r1", string(app.RunnerJobGroupSandbox),
		"sandbox-terraform", "apply-plan", "install_sandbox_runs", "run00000000000000000000001", base.Add(time.Minute), "")
	seedRunnerJob(t, db, "job00000000000000000000003", "r1", string(app.RunnerJobGroupBuild),
		"docker-build", "build", "builds000000000000000000001", "bld00000000000000000000001", base.Add(2*time.Minute), "")

	steps, err := exportInstallSteps(context.Background(), db, false, installID)
	require.NoError(t, err)
	require.Len(t, steps, 2)

	require.Equal(t, "job00000000000000000000001", steps[0].ID)
	require.Equal(t, "create-apply-plan", steps[0].JobOperation)
	require.Empty(t, steps[0].DependsOn)

	require.Equal(t, "job00000000000000000000002", steps[1].ID)
	require.Equal(t, "apply-plan", steps[1].JobOperation)
	require.Equal(t, []string{"job00000000000000000000001"}, steps[1].DependsOn)
	require.Equal(t, "job00000000000000000000001", steps[1].PlanFromStep)
}

func TestExportInstallStepsNoJobsReturnsNil(t *testing.T) {
	db := planExportTestDB(t)
	steps, err := exportInstallSteps(context.Background(), db, false, "inl00000000000000000000404")
	require.NoError(t, err)
	require.Nil(t, steps)
}

func componentConnectionsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)

	for _, ddl := range []string{
		`CREATE TABLE app_configs (
			id text primary key, created_at datetime, updated_at datetime, deleted_at integer not null default 0,
			component_ids text
		)`,
		`CREATE TABLE components (id text primary key, name text, deleted_at integer not null default 0)`,
		`CREATE TABLE component_config_connections (
			id text primary key, created_at datetime, updated_at datetime, deleted_at integer not null default 0,
			app_config_id text, component_id text, component_dependency_ids text, "references" text
		)`,
		`CREATE TABLE terraform_module_component_configs (
			id text primary key, created_at datetime, updated_at datetime, deleted_at integer not null default 0,
			component_config_connection_id text
		)`,
		`CREATE TABLE helm_component_configs (
			id text primary key, created_at datetime, updated_at datetime, deleted_at integer not null default 0,
			component_config_connection_id text
		)`,
		`CREATE TABLE docker_build_component_configs (
			id text primary key, deleted_at integer not null default 0, component_config_connection_id text
		)`,
		`CREATE TABLE external_image_component_configs (
			id text primary key, deleted_at integer not null default 0, component_config_connection_id text
		)`,
		`CREATE TABLE job_component_configs (
			id text primary key, deleted_at integer not null default 0, component_config_connection_id text
		)`,
		`CREATE TABLE kubernetes_manifest_component_configs (
			id text primary key, deleted_at integer not null default 0, component_config_connection_id text
		)`,
		`CREATE TABLE pulumi_component_configs (
			id text primary key, deleted_at integer not null default 0, component_config_connection_id text
		)`,
		`CREATE TABLE public_git_vcs_configs (
			id text primary key, deleted_at integer not null default 0, component_config_id text, component_config_type text
		)`,
		`CREATE TABLE connected_github_vcs_configs (
			id text primary key, deleted_at integer not null default 0, component_config_id text, component_config_type text
		)`,
		`CREATE VIEW component_config_connections_latest_configs_view AS SELECT * FROM component_config_connections`,
	} {
		require.NoError(t, db.Exec(ddl).Error)
	}
	return db
}

func TestExportComponentConfigConnectionsNestedConfigsAndLatestFallback(t *testing.T) {
	db := componentConnectionsTestDB(t)

	// cfg-2 references both components, but only comp-a's connection is pinned
	// to cfg-2; comp-b was unchanged in this version, so its config only exists
	// on cfg-1 and must be found through the latest-configs view.
	require.NoError(t, db.Exec(`INSERT INTO app_configs (id, component_ids) VALUES ('apc00000000000000000000002', '{cmp0000000000000000000000a,cmp0000000000000000000000b}')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO components (id, name) VALUES ('cmp0000000000000000000000a', 'certificate'), ('cmp0000000000000000000000b', 'application_load_balancer')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO component_config_connections (id, app_config_id, component_id) VALUES
		('ccc0000000000000000000000a', 'apc00000000000000000000002', 'cmp0000000000000000000000a'),
		('ccc0000000000000000000000b', 'apc00000000000000000000001', 'cmp0000000000000000000000b')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO terraform_module_component_configs (id, component_config_connection_id) VALUES ('tfc0000000000000000000000a', 'ccc0000000000000000000000a')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO helm_component_configs (id, component_config_connection_id) VALUES ('hlm0000000000000000000000b', 'ccc0000000000000000000000b')`).Error)

	connections, err := exportComponentConfigConnections(context.Background(), db, "apc00000000000000000000002")
	require.NoError(t, err)
	require.Len(t, connections, 2)

	byComponent := map[string]app.ComponentConfigConnection{}
	for _, connection := range connections {
		byComponent[connection.ComponentID] = connection
	}
	require.NotNil(t, byComponent["cmp0000000000000000000000a"].TerraformModuleComponentConfig, "pinned connection must carry its terraform config")
	require.NotNil(t, byComponent["cmp0000000000000000000000b"].HelmComponentConfig, "latest-view fallback connection must carry its helm config")
}

func TestExportComponentConfigConnectionsMissingConfigFails(t *testing.T) {
	db := componentConnectionsTestDB(t)

	require.NoError(t, db.Exec(`INSERT INTO app_configs (id, component_ids) VALUES ('apc00000000000000000000003', '{cmp0000000000000000000000z}')`).Error)

	_, err := exportComponentConfigConnections(context.Background(), db, "apc00000000000000000000003")
	require.ErrorContains(t, err, "found 0 component configs, expected 1")
}

func TestExportInstallStepsDeduplicatesRetriedSandboxRuns(t *testing.T) {
	db := planExportTestDB(t)
	installID := "inl00000000000000000000002"

	require.NoError(t, db.Exec(`INSERT INTO runner_groups (id, owner_id, owner_type) VALUES ('rg1', ?, 'installs')`, installID).Error)
	require.NoError(t, db.Exec(`INSERT INTO runners (id, runner_group_id) VALUES ('r1', 'rg1')`).Error)

	base := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	sandboxPlan := `{"sandbox_run_plan":{}}`
	seedRunnerJob(t, db, "job00000000000000000000011", "r1", string(app.RunnerJobGroupSandbox),
		"sandbox-terraform", "create-apply-plan", "install_sandbox_runs", "runold0000000000000000001", base, sandboxPlan)
	seedRunnerJob(t, db, "job00000000000000000000012", "r1", string(app.RunnerJobGroupSandbox),
		"sandbox-terraform", "create-apply-plan", "install_sandbox_runs", "runnew0000000000000000002", base.Add(time.Hour), sandboxPlan)

	steps, err := exportInstallSteps(context.Background(), db, false, installID)
	require.NoError(t, err)
	require.Len(t, steps, 1)
	require.Equal(t, "job00000000000000000000012", steps[0].ID)
}
