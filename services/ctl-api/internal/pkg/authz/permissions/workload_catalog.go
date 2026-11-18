package permissions

const (
	ExecutionConfigRead   PermissionID = "execution.config.read"
	ActionRunReport       PermissionID = "action.run.report"
	StackConfigRead       PermissionID = "stack.config.read"
	StackVersionReport    PermissionID = "stack.version.report"
	TelemetrySettingsRead PermissionID = "telemetry.settings.read"
	TelemetryTokenCreate  PermissionID = "telemetry.token.create"
)

func WorkloadDefinitions() []PermissionDefinition {
	install := []ResourceKind{KindOrg, KindApp, KindInstall}
	return []PermissionDefinition{
		{ID: ExecutionConfigRead, Title: "Read execution configuration", Description: "Read full app and action configuration on the runner API. Install-scoped workload callers can read configurations from their verified runner's app.", Scopes: install, Targets: []ResourceKind{KindOperation}},
		{ID: ActionRunReport, Title: "Report action execution", Description: "Update status and duration on an action run step under its verified install and run.", Scopes: install, Targets: []ResourceKind{KindOperation}},
		{ID: StackConfigRead, Title: "Read stack configuration", Description: "Read the full configuration used by an install's stack.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: StackReport, Title: "Report stack outputs", Description: "Submit stack phone-home results for an install using the existing version and callback validation.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: StackVersionReport, Title: "Report stack version outputs", Description: "Authenticate the legacy version-specific phone-home callback, subject to its credential and version validation.", Scopes: []ResourceKind{KindOrg, KindApp, KindInstall, KindPhoneHome}, Targets: []ResourceKind{KindPhoneHome}},
		{ID: TelemetrySettingsRead, Title: "Read telemetry settings", Description: "Read an install's collector settings without runner credentials or backend secrets.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: TelemetryTokenCreate, Title: "Issue telemetry credentials", Description: "Issue an install-scoped telemetry relay credential with the existing audience and expiry.", Scopes: install, Targets: []ResourceKind{KindInstall}},
	}
}
