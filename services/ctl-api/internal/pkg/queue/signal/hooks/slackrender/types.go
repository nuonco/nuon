// Package slackrender renders Slack message text and blocks for Nuon
// workflow / workflow_step / workflow_step_approval lifecycle events.
//
// The renderer is purely a pure function of its input Event — it never
// calls back into ctl-api. Producers (the slack lifecycle hook) are
// expected to enrich the Event from the same canonical sources the
// webhook hook uses.
//
// This package mirrors the input shape used by slackbot's notify package
// (lifecycle.WebhookPayload / EventData / WorkflowRef / StepRef / etc.)
// so the rendering logic can be ported with minimal adaptation. The
// difference from slackbot is the OUTPUT contract: slackbot emits one
// inlined line; we keep the existing two-block shape — a SECTION block
// (bold headline + optional status sub-line) plus a CONTEXT block
// (small grey footer with org/install/workflow chips and an Open in Nuon
// link).
package slackrender

import "time"

// Kind values for Event.Kind. Matches the wire vocabulary in
// services/ctl-api/internal/pkg/queue/signal/hooks/webhook.go.
const (
	KindWorkflow             = "workflow"
	KindWorkflowStep         = "workflow_step"
	KindWorkflowStepApproval = "workflow_step_approval"
)

// Transition values for Event.Transition. British "cancelled" is the
// canonical wire value (matches webhook.go's transitionCanceled).
const (
	TransitionStarted   = "started"
	TransitionSucceeded = "succeeded"
	TransitionFailed    = "failed"
	TransitionCancelled = "cancelled"

	TransitionRequested = "requested"
	TransitionApproved  = "approved"
	TransitionRejected  = "rejected"
)

// OwnerType values for WorkflowRef.OwnerType.
const (
	OwnerTypeInstalls    = "installs"
	OwnerTypeApps        = "apps"
	OwnerTypeAppBranches = "app_branches"
)

// WorkflowType values for WorkflowRef.Type.
const (
	WorkflowTypeProvision          = "provision"
	WorkflowTypeReprovision        = "reprovision"
	WorkflowTypeManualDeploy       = "manual_deploy"
	WorkflowTypeActionWorkflowRun  = "action_workflow_run"
	WorkflowTypeDriftRun           = "drift_run"
	WorkflowTypeDeployComponents   = "deploy_components"
	WorkflowTypeTeardownComponent  = "teardown_component"
	WorkflowTypeTeardownComponents = "teardown_components"
	WorkflowTypeInputUpdate        = "input_update"
	WorkflowTypeSyncSecrets        = "sync_secrets"
	WorkflowTypeDeprovision        = "deprovision"
	WorkflowTypeDeprovisionSandbox = "deprovision_sandbox"
	WorkflowTypeReprovisionSandbox = "reprovision_sandbox"
	WorkflowTypeAppConfigBuild     = "app_config_build"
	WorkflowTypeAppBranchesRun     = "app_branches_manual_update"
)

// TargetType values for StepRef.TargetType. Matches the actual string
// values stored in install_workflow_steps.step_target_type, which is the
// plural form of the app.WorkflowStepTargetType* constants — the singular
// constants exist but are not what the DB carries.
const (
	TargetTypeInstallDeploys             = "install_deploys"
	TargetTypeInstallSandboxRuns         = "install_sandbox_runs"
	TargetTypeInstallActionWorkflowRuns  = "install_action_workflow_runs"
	TargetTypeInstallCloudFormationStack = "install_cloudformation_stack"
	TargetTypeInstallRunnerUpdate        = "install_runner_update"
)

// WorkflowRef identifies the workflow this event is about.
type WorkflowRef struct {
	ID        string
	Type      string
	OwnerID   string
	OwnerType string
	OwnerName string
	// CreatedByEmail labels who started the workflow. Falls back to the
	// raw account id for accounts without an email; empty when the
	// workflow has no creator.
	CreatedByEmail string
	// CreatedAt is the workflow's start time. Zero when unknown.
	CreatedAt time.Time
}

// StepRef identifies a workflow step. Present only on workflow_step /
// workflow_step_approval events.
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

// ParentRef is set when this workflow was launched from another workflow's
// step (e.g. an action workflow run launched from a deploy step).
type ParentRef struct {
	WorkflowID string
	StepID     string
	Kind       string
	ActionName string
}

// Outcome is set on terminal transitions (succeeded / failed / cancelled).
type Outcome struct {
	Status     string
	Error      string
	DurationMs int64
}

// ApprovalRef carries the structured approval block on
// workflow_step_approval events.
type ApprovalRef struct {
	ID          string
	Type        string
	Plan        string
	RespondedBy string
}

// ContextLinks contains dashboard / API URLs for the entities referenced
// in the event. Any field may be empty.
type ContextLinks struct {
	Org        string
	Install    string
	Workflow   string
	Sandbox    string
	Component  string
	Approval   string
	RespondAPI string
}

// Event is the input to every Build* renderer function. It mirrors the
// ctl-api webhook payload shape (lifecycleEventData) one-to-one so the
// slack lifecycle hook can build it from the same enrichment sources.
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
}

// IsTerminal reports whether the event represents a terminal transition.
func (e Event) IsTerminal() bool {
	switch e.Transition {
	case TransitionSucceeded, TransitionFailed, TransitionCancelled,
		TransitionApproved, TransitionRejected:
		return true
	}
	return false
}
