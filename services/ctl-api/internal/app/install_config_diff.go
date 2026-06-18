package app

type ComponentDiffEntry struct {
	ComponentID   string `json:"component_id"`
	ComponentName string `json:"component_name,omitempty"`
	OldChecksum   string `json:"old_checksum,omitempty"`
	NewChecksum   string `json:"new_checksum,omitempty"`
}

type InstallConfigDiff struct {
	Added     []ComponentDiffEntry `json:"added"`
	Removed   []ComponentDiffEntry `json:"removed"`
	Changed   []ComponentDiffEntry `json:"changed"`
	Unchanged []ComponentDiffEntry `json:"unchanged"`

	SandboxChanged bool   `json:"sandbox_changed"`
	SandboxOldID   string `json:"sandbox_old_id,omitempty"`
	SandboxNewID   string `json:"sandbox_new_id,omitempty"`

	StackChanged bool   `json:"stack_changed"`
	StackOldID   string `json:"stack_old_id,omitempty"`
	StackNewID   string `json:"stack_new_id,omitempty"`
}
