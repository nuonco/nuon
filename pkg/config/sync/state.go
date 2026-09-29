package sync

const (
	DefaultStateVersion string = "v1"
)

type ActionState struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type RunbookState struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type AppBranchConfigState struct {
	AppBranchID       string `json:"app_branch_id"`
	AppBranchConfigID string `json:"app_branch_config_id"`
}

type State struct {
	Version string `json:"version"`

	CfgID           string           `json:"config_id"`
	AppID           string           `json:"app_id"`
	InstallerID     string           `json:"installer_id"`
	RunnerConfigID  string           `json:"runner_config_id"`
	SandboxConfigID string           `json:"sandbox_config_id"`
	InputConfigID   string           `json:"input_config_id"`
	Components      []ComponentState `json:"components"`
	Actions         []ActionState    `json:"actions"`
	Runbooks        []RunbookState   `json:"runbooks"`

	Result *Result `json:"result,omitempty"`
}

type Result struct {
	ComponentsScheduled []ComponentState `json:"components_scheduled,omitempty"`

	ComponentsCreated []string `json:"components_created,omitempty"`

	AppBranchesCreated []string `json:"app_branches_created,omitempty"`

	AppBranchConfigsUpdated []AppBranchConfigState `json:"app_branch_configs_updated,omitempty"`

	OrphanedComponents map[string]string `json:"orphaned_components,omitempty"`
	OrphanedActions    map[string]string `json:"orphaned_actions,omitempty"`
	OrphanedRunbooks   map[string]string `json:"orphaned_runbooks,omitempty"`
}
