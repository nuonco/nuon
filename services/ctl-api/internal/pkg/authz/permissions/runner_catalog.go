package permissions

const (
	RunnerInspect        PermissionID = "runner.inspect"
	RunnerControl        PermissionID = "runner.control"
	RunnerSettingsUpdate PermissionID = "runner.settings.update"
	RunnerTokensPrune    PermissionID = "runner.tokens.prune"
	TerraformRead        PermissionID = "terraform.read"
	TerraformCreate      PermissionID = "terraform.create"
	TerraformDelete      PermissionID = "terraform.delete"
	TerraformLock        PermissionID = "terraform.lock"
	TerraformStateWrite  PermissionID = "terraform.state.write"
)

func RunnerDefinitions() []PermissionDefinition {
	scopes := []ResourceKind{KindOrg, KindApp, KindInstall}
	targets := []ResourceKind{KindRunner}
	return []PermissionDefinition{
		{ID: RunnerInspect, Title: "Inspect runners", Description: "Read runners, settings, processes, health and heartbeats under the runner group's verified org or install owner.", Scopes: scopes, Targets: targets},
		{ID: RunnerControl, Title: "Control runners", Description: "Shut down processes and manage runner instances, including restart, update and VM shutdown.", Scopes: scopes, Targets: targets},
		{ID: RunnerSettingsUpdate, Title: "Update runner settings", Description: "Update the owned runner group's settings, affecting its current and future runners.", Scopes: scopes, Targets: targets},
		{ID: RunnerTokensPrune, Title: "Prune runner credentials", Description: "Revoke old credentials for an owned runner using the existing token retention rules.", Scopes: scopes, Targets: targets},
		{ID: TerraformRead, Title: "Read Terraform state", Description: "Read workspaces, locks and complete state history, including raw Terraform and JSON/Pulumi state, under verified owners.", Scopes: scopes, Targets: []ResourceKind{KindOperation}},
		{ID: TerraformCreate, Title: "Create Terraform workspaces", Description: "Create a workspace for a verified owner.", Scopes: scopes, Targets: []ResourceKind{KindOperation}},
		{ID: TerraformDelete, Title: "Delete Terraform workspaces", Description: "Delete a workspace and its existing state cleanup.", Scopes: scopes, Targets: []ResourceKind{KindOperation}},
		{ID: TerraformLock, Title: "Lock Terraform workspaces", Description: "Lock and unlock workspaces using the existing lock ownership contract.", Scopes: scopes, Targets: []ResourceKind{KindOperation}},
		{ID: TerraformStateWrite, Title: "Write Terraform state", Description: "Write raw Terraform backend state or JSON/Pulumi state to an owned workspace.", Scopes: scopes, Targets: []ResourceKind{KindOperation}},
	}
}
