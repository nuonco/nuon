package app

import (
	"encoding/json"
	"fmt"
)

const RunbookTargetsMetadataKey = "runbook_targets_v1"
const ManualDeployTargetsMetadataKey = "manual_deploy_targets_v1"

type InstallExecutionTargets struct {
	OrgID              string                            `json:"org_id"`
	AppID              string                            `json:"app_id"`
	InstallID          string                            `json:"install_id"`
	AppConfigID        string                            `json:"app_config_id"`
	RunbookConfigID    string                            `json:"runbook_config_id"`
	Components         map[string]RunbookComponentTarget `json:"components"`
	InstallComponents  map[string]string                 `json:"install_components"`
	TeardownComponents map[string]RunbookTeardownTarget  `json:"teardown_components"`
	Actions            map[string]RunbookActionTarget    `json:"actions"`
	StepComponents     map[string][]string               `json:"step_components"`
	ComponentOrder     []string                          `json:"component_order"`
	InstallSandboxID   string                            `json:"install_sandbox_id,omitempty"`
	PrimaryComponentID string                            `json:"primary_component_id,omitempty"`
	PrimaryBuildID     string                            `json:"primary_build_id,omitempty"`
	PrimaryType        ComponentType                     `json:"primary_type,omitempty"`
}

type RunbookComponentTarget struct {
	ConfigID string `json:"config_id"`
	Name     string `json:"name"`
}

type RunbookActionTarget struct {
	InstallActionWorkflowID string `json:"install_action_workflow_id"`
	ConfigID                string `json:"config_id"`
}

type RunbookTeardownTarget struct {
	Type     ComponentType `json:"type"`
	Name     string        `json:"name"`
	ConfigID string        `json:"config_id,omitempty"`
}

func (w *Workflow) ExecutionTargets() (*InstallExecutionTargets, error) {
	value, ok := w.Metadata[RunbookTargetsMetadataKey]
	typ := WorkflowTypeRunbookRun
	if !ok {
		value, ok = w.Metadata[ManualDeployTargetsMetadataKey]
		typ = WorkflowTypeManualDeploy
	}
	if !ok {
		return nil, nil
	}
	if value == nil {
		return nil, fmt.Errorf("missing execution targets")
	}
	var targets InstallExecutionTargets
	if err := json.Unmarshal([]byte(*value), &targets); err != nil {
		return nil, fmt.Errorf("invalid execution targets: %w", err)
	}
	if w.Type != typ || w.OwnerType != "installs" || targets.InstallID != w.OwnerID || targets.OrgID != w.OrgID || targets.AppConfigID == "" {
		return nil, fmt.Errorf("execution targets do not belong to workflow")
	}
	if typ == WorkflowTypeManualDeploy && (targets.PrimaryComponentID == "" || targets.PrimaryBuildID == "") {
		return nil, fmt.Errorf("missing manual deploy primary target")
	}
	return &targets, nil
}

func (t *InstallExecutionTargets) CheckInstall(install *Install) error {
	if install == nil || t.InstallID != install.ID || t.OrgID != install.OrgID || t.AppID != install.AppID || t.AppConfigID != install.AppConfigID {
		return fmt.Errorf("install or app config changed after the operation was requested; start a new operation")
	}
	return nil
}

func (t *InstallExecutionTargets) CheckComponent(componentID string, component *InstallComponent) error {
	if component == nil || t.InstallComponents[componentID] == "" || component.ID != t.InstallComponents[componentID] || component.ComponentID != componentID || component.InstallID != t.InstallID || component.OrgID != t.OrgID {
		return fmt.Errorf("installed component does not match prepared runbook target")
	}
	return nil
}

func (w *Workflow) PreparedInstallConfigID(install *Install) (string, error) {
	targets, err := w.ExecutionTargets()
	if err != nil {
		return "", err
	}
	if targets != nil {
		return targets.AppConfigID, targets.CheckInstall(install)
	}
	if w.Type != WorkflowTypeActionWorkflowRun {
		return "", nil
	}
	if value, ok := w.Metadata["manual_action_app_config_id"]; ok {
		if install == nil || value == nil || *value == "" || *value != install.AppConfigID || w.OrgID != install.OrgID || w.OwnerID != install.ID || w.OwnerType != "installs" {
			return "", fmt.Errorf("install or app config changed after the operation was requested; start a new operation")
		}
		return *value, nil
	}
	return "", nil
}
