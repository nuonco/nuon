package slackrender

import "time"

const (
	KindWorkflow             = "workflow"
	KindWorkflowStep         = "workflow_step"
	KindWorkflowStepApproval = "workflow_step_approval"
	KindRoleChange           = "role_change"
	KindInputsUpdated        = "inputs_updated"
	KindStackRun             = "stack_run"
	KindAppConfigSynced      = "app_config_synced"
	KindRunnerUnhealthy      = "runner_unhealthy"
	KindLabelAdded           = "label_added"
	KindAppBranchChanged     = "app_branch_changed"
)

const (
	TransitionStarted   = "started"
	TransitionSucceeded = "succeeded"
	TransitionFailed    = "failed"
	TransitionCancelled = "cancelled"

	TransitionRequested     = "requested"
	TransitionApproved      = "approved"
	TransitionRejected      = "rejected"
	TransitionAwaitingRetry = "awaiting_retry"
	TransitionUnhealthy     = "unhealthy"
)

const (
	OwnerTypeInstalls    = "installs"
	OwnerTypeApps        = "apps"
	OwnerTypeAppBranches = "app_branches"
)

const (
	WorkflowTypeProvision                      = "provision"
	WorkflowTypeReprovision                    = "reprovision"
	WorkflowTypeManualDeploy                   = "manual_deploy"
	WorkflowTypeActionWorkflowRun              = "action_workflow_run"
	WorkflowTypeDriftRun                       = "drift_run"
	WorkflowTypeDeployComponents               = "deploy_components"
	WorkflowTypeTeardownComponent              = "teardown_component"
	WorkflowTypeTeardownComponents             = "teardown_components"
	WorkflowTypeInputUpdate                    = "input_update"
	WorkflowTypeSyncSecrets                    = "sync_secrets"
	WorkflowTypeDeprovision                    = "deprovision"
	WorkflowTypeDeprovisionSandbox             = "deprovision_sandbox"
	WorkflowTypeReprovisionSandbox             = "reprovision_sandbox"
	WorkflowTypeReprovisionStack               = "reprovision_stack"
	WorkflowTypeAppConfigBuild                 = "app_config_build"
	WorkflowTypeAppBranchesRun                 = "app_branches_manual_update"
	WorkflowTypeAppBranchesConfigRepoUpdate    = "app_branches_config_repo_update"
	WorkflowTypeAppBranchesComponentRepoUpdate = "app_branches_component_repo_update"
	WorkflowTypeRunbookRun                     = "runbook_run"
	WorkflowTypeComponentEnabled               = "component_enabled"
	WorkflowTypeComponentDisabled              = "component_disabled"
)

const (
	TargetTypeInstallDeploys             = "install_deploys"
	TargetTypeInstallSandboxRuns         = "install_sandbox_runs"
	TargetTypeInstallActionWorkflowRuns  = "install_action_workflow_runs"
	TargetTypeInstallCloudFormationStack = "install_cloudformation_stack"
	TargetTypeInstallRunnerUpdate        = "install_runner_update"
)

type WorkflowRef struct {
	ID             string
	Type           string
	OwnerID        string
	OwnerType      string
	OwnerName      string
	CreatedByEmail string
	CreatedAt      time.Time
	RunbookName    string
}

type StepRef struct {
	ID            string
	Name          string
	Idx           int
	TargetType    string
	TargetID      string
	ComponentID   string
	ComponentName string
	SandboxID     string
	ExecutionType string
}

type ParentRef struct {
	WorkflowID string
	StepID     string
	Kind       string
	ActionName string
}

type Outcome struct {
	Status     string
	Error      string
	DurationMs int64
}

type ApprovalRef struct {
	ID          string
	Type        string
	Plan        string
	RespondedBy string
}

type ContextLinks struct {
	Org        string
	Install    string
	Workflow   string
	Sandbox    string
	Component  string
	Approval   string
	RespondAPI string
}

type Event struct {
	Kind       string
	Transition string

	OrgID   string
	OrgName string

	Workflow WorkflowRef
	Step     *StepRef
	Parent   *ParentRef
	Outcome  *Outcome
	Approval *ApprovalRef
	Links    *ContextLinks

	Metadata map[string]any
}

func (e Event) IsTerminal() bool {
	switch e.Transition {
	case TransitionSucceeded, TransitionFailed, TransitionCancelled,
		TransitionApproved, TransitionRejected:
		return true
	}
	return false
}
