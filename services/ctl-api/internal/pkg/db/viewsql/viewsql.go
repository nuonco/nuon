package viewsql

import (
	_ "embed"
)

//go:embed action_workflow_configs_latest_view_v1.sql
var ActionWorkflowConfigsViewV1 string

//go:embed installs_view_v3.sql
var InstallsViewV3 string

//go:embed installs_view_v4.sql
var InstallsViewV4 string

//go:embed installs_view_v5.sql
var InstallsViewV5 string

//go:embed installs_view_v6.sql
var InstallsViewV6 string

//go:embed install_states_view_v1.sql
var InstallStatesViewV1 string

//go:embed app_configs_view_v2.sql
var AppConfigViewV2 string

//go:embed component_config_connections_v1.sql
var ComponentConfigConnectionsV1 string

//go:embed latest_component_config_connections_view_v1.sql
var LatestComponentConfigConnectionsV1 string

//go:embed install_action_workflow_latest_runs_view_v1.sql
var InstallActionWorkflowLatestRunsViewV1 string

//go:embed install_action_workflow_runs_state_view_v1.sql
var InstallActionWorkflowRunsStateViewV1 string

//go:embed install_sandbox_runs_state_view_v1.sql
var InstallSandboxRunsStateViewV1 string

//go:embed install_deploys_state_view_v1.sql
var InstallDeploysStateViewV1 string

//go:embed install_stack_version_runs_state_view_v1.sql
var InstallStackVersionRunsStateViewV1 string

//go:embed install_deploys_latest_view_v1.sql
var InstallDeploysLatestViewV1 string

//go:embed install_inputs_view_v1.sql
var InstallInputsViewV1 string

//go:embed runner_health_checks_view_v1.sql
var RunnerHealthCheckViewV1 string

//go:embed runner_jobs_view_v1.sql
var RunnerJobViewV1 string

//go:embed runner_jobs_view_v2.sql
var RunnerJobViewV2 string

//go:embed runner_settings_v1.sql
var RunnerSettingsV1 string

//go:embed runner_wide_v1.sql
var RunnerWideV1 string

//go:embed psql_table_sizes_v1.sql
var PSQLTableSizesV1 string

//go:embed install_audit_logs_view_v1.sql
var InstallAuditLogsViewV1 string

//go:embed ch_table_sizes_v1.sql
var CHTableSizesV1 string

//go:embed terraform_workspace_states_view_v1.sql
var TerraformWorkspaceStatesViewV1 string
