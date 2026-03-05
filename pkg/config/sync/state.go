package sync

const (
	// DefaultStateVersion is the current version of the state format
	DefaultStateVersion string = "v1"
)

// ActionState represents an action in the sync state
type ActionState struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// State represents the synchronized state of an app config.
// This is stored as JSON in the app_configs.state column to track
// what was synced in each config version.
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
}
