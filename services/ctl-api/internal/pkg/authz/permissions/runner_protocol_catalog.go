package permissions

const (
	RunnerHealthReport          PermissionID = "runner.health.report"
	RunnerComponentHealthRead   PermissionID = "runner.component_health.read"
	RunnerComponentHealthReport PermissionID = "runner.component_health.report"
	RunnerQueueRead             PermissionID = "runner.queue.read"
	RunnerProcessReport         PermissionID = "runner.process.report"
	RunnerTelemetryWrite        PermissionID = "runner.telemetry.write"
	RunnerTelemetryTokenCreate  PermissionID = "runner.telemetry_token.create"
	RunnerSandboxConfigRead     PermissionID = "runner.sandbox_config.read"
	RunnerJobReport             PermissionID = "runner_job.report"
	TerraformStateDelete        PermissionID = "terraform.state.delete"
	HelmReleaseRead             PermissionID = "helm.release.read"
	HelmReleaseWrite            PermissionID = "helm.release.write"
	HelmReleaseDelete           PermissionID = "helm.release.delete"
)

func RunnerProtocolDefinitions() []PermissionDefinition {
	scopes := []ResourceKind{KindOrg, KindApp, KindInstall}
	runnerScopes := []ResourceKind{KindOrg, KindApp, KindInstall, KindRunner}
	jobScopes := []ResourceKind{KindOrg, KindApp, KindInstall, KindRunnerJob}
	runner := []ResourceKind{KindRunner}
	job := []ResourceKind{KindRunnerJob}
	state := []ResourceKind{KindOperation}
	return []PermissionDefinition{
		{ID: RunnerHealthReport, Title: "Report runner health", Description: "Submit runner health checks and heartbeats, including process liveness and credential bookkeeping.", Scopes: runnerScopes, Targets: runner},
		{ID: RunnerComponentHealthRead, Title: "Read component health configuration", Description: "Read the runner install's observable components and persisted cluster access context.", Scopes: runnerScopes, Targets: runner},
		{ID: RunnerComponentHealthReport, Title: "Report component health", Description: "Submit component health and persist cluster access context for the runner's verified install.", Scopes: runnerScopes, Targets: runner},
		{ID: RunnerQueueRead, Title: "Poll runner queue", Description: "Enumerate and long-poll all available jobs assigned to this runner, including jobs across apps on an org-owned build runner. This is not filtered customer job inspection.", Scopes: runnerScopes, Targets: runner},
		{ID: RunnerProcessReport, Title: "Report runner processes", Description: "Register processes and their log streams, report lifecycle status and termination, and acknowledge shutdown requests. Does not initiate customer shutdown commands.", Scopes: runnerScopes, Targets: runner},
		{ID: RunnerTelemetryWrite, Title: "Submit runner telemetry", Description: "Ingest runner OTLP traces and metrics.", Scopes: runnerScopes, Targets: runner},
		{ID: RunnerTelemetryTokenCreate, Title: "Issue runner telemetry token", Description: "Issue the existing short-lived telemetry relay credential for an eligible install-runner service account.", Scopes: runnerScopes, Targets: runner},
		{ID: RunnerSandboxConfigRead, Title: "Read runner sandbox configuration", Description: "Load shared sandbox-mode execution templates through a verified runner. Templates are global, not private to the runner's app or install.", Scopes: runnerScopes, Targets: runner},
		{ID: RunnerJobClaim, Title: "Claim runner job credential", Description: "Exchange an owning runner credential for an available job's execution token, preserving any credential ceiling and the existing subject and single-claim guards. Bounded claims require grant-based job access.", Scopes: jobScopes, Targets: job},
		{ID: RunnerJobReport, Title: "Report runner job execution", Description: "Start job execution and report status, results and outputs using the job's own credential. Does not grant customer cancellation or runner control.", Scopes: jobScopes, Targets: job},
		{ID: TerraformStateDelete, Title: "Delete workspace state", Description: "Delete stored workspace JSON state or invoke the legacy no-op Terraform backend deletion endpoint.", Scopes: scopes, Targets: state},
		{ID: HelmReleaseRead, Title: "Read Helm release state", Description: "Read, list and query complete Helm releases under a chart's verified owner.", Scopes: scopes, Targets: state},
		{ID: HelmReleaseWrite, Title: "Write Helm release state", Description: "Create or update stored Helm release bodies and labels under a chart's verified owner.", Scopes: scopes, Targets: state},
		{ID: HelmReleaseDelete, Title: "Delete Helm release state", Description: "Delete stored Helm releases under a chart's verified owner.", Scopes: scopes, Targets: state},
	}
}
