package app

import configdiff "github.com/nuonco/nuon/pkg/config/diff"

type InstallConfigImpact string

const (
	InstallConfigImpactStackConfig  InstallConfigImpact = "stack_config"
	InstallConfigImpactPermissions  InstallConfigImpact = "permissions"
	InstallConfigImpactBreakGlass   InstallConfigImpact = "break_glass"
	InstallConfigImpactSecrets      InstallConfigImpact = "secrets"
	InstallConfigImpactRunnerConfig InstallConfigImpact = "runner_config"
)

type ComponentDiffEntry struct {
	ComponentID   string                    `json:"component_id"`
	ComponentName string                    `json:"component_name,omitempty"`
	ComponentType string                    `json:"component_type,omitempty"`
	OldChecksum   string                    `json:"old_checksum,omitempty"`
	NewChecksum   string                    `json:"new_checksum,omitempty"`
	OldBuildID    string                    `json:"old_build_id,omitempty"`
	NewBuildID    string                    `json:"new_build_id,omitempty"`
	BuildChanged  bool                      `json:"build_changed,omitempty"`
	ImpactReasons []configdiff.ImpactReason `json:"impact_reasons,omitempty"`
}

type InstallConfigDiff struct {
	Added     []ComponentDiffEntry `json:"added"`
	Removed   []ComponentDiffEntry `json:"removed"`
	Changed   []ComponentDiffEntry `json:"changed"`
	Unchanged []ComponentDiffEntry `json:"unchanged"`

	SandboxChanged bool   `json:"sandbox_changed"`
	SandboxOldID   string `json:"sandbox_old_id,omitempty"`
	SandboxNewID   string `json:"sandbox_new_id,omitempty"`

	SandboxBuildChanged bool   `json:"sandbox_build_changed"`
	SandboxBuildOldID   string `json:"sandbox_build_old_id,omitempty"`
	SandboxBuildNewID   string `json:"sandbox_build_new_id,omitempty"`

	StackChanged       bool                      `json:"stack_changed"`
	StackOldID         string                    `json:"stack_old_id,omitempty"`
	StackNewID         string                    `json:"stack_new_id,omitempty"`
	StackImpacts       []InstallConfigImpact     `json:"stack_impacts,omitempty"`
	StackImpactReasons []configdiff.ImpactReason `json:"stack_impact_reasons,omitempty"`
}
