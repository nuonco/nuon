package signals

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	TemporalNamespace string = "installs"
)

// SignalType is a string identifier for signal operations.
type SignalType = string

const (
	OperationForget                      SignalType = "forgotten"
	OperationRestart                     SignalType = "restart"
	OperationRestartChildren             SignalType = "restart-children"
	OperationSyncActionWorkflowTriggers  SignalType = "sync-action-workflow-triggers"
	OperationActionWorkflowRun           SignalType = "action-workflow-run"
	OperationPollDependencies            SignalType = "poll-dependencies"
	OperationCreated                     SignalType = "created"
	OperationUpdated                     SignalType = "updated"
	OperationGenerateInstallStackVersion SignalType = "generate-install-stack-version"
	OperationProvisionRunner             SignalType = "provision-runner"
	OperationAwaitInstallStackVersionRun SignalType = "await-install-stack-version-run"
	OperationUpdateInstallStackOutputs   SignalType = "update-install-stack-outputs"
	OperationAwaitRunnerHealthy          SignalType = "await-runner-healthy"
	OperationProvisionSandbox            SignalType = "provision-sandbox"
	OperationProvisionDNS                SignalType = "provision-dns"
	OperationDeprovisionDNS              SignalType = "deprovision-dns"
	OperationExecuteActionWorkflow       SignalType = "execute-action-workflow"
	OperationExecuteDeployComponent      SignalType = "execute-deploy-component"
	OperationExecuteTeardownComponent    SignalType = "execute-teardown-component"
	OperationSyncSecrets                 SignalType = "sync-secrets"
	OperationWorkflowApproveAll          SignalType = "workflow-approve-all"
	OperationWorkflowStepApprovalReq     SignalType = "workflow-step-approval-request"
	OperationWorkflowStepApprovalResp    SignalType = "workflow-step-approval-response"
	OperationGenerateState               SignalType = "generate-state"

	OperationExecuteFlow SignalType = "execute-workflow"
	OperationRerunFlow   SignalType = "rerun-flow"

	OperationExecuteDeployComponentSyncImage SignalType = "component-sync-image"

	OperationExecuteDeployComponentSyncAndPlan SignalType = "component-deploy-sync-and-plan"
	OperationExecuteDeployComponentApplyPlan   SignalType = "component-deploy-apply-plan"
	OperationExecuteDeployComponentPlanOnly    SignalType = "component-deploy-plan-only"

	OperationExecuteTeardownComponentSyncAndPlan SignalType = "component-teardown-sync-and-plan"
	OperationExecuteTeardownComponentApplyPlan   SignalType = "component-teardown-apply-plan"

	OperationProvisionSandboxPlan        SignalType = "provision-sandbox-plan"
	OperationProvisionSandboxApplyPlan   SignalType = "provision-sandbox-apply-plan"
	OperationDeprovisionSandboxPlan      SignalType = "deprovision-sandbox-plan"
	OperationDeprovisionSandboxApplyPlan SignalType = "deprovision-sandbox-apply-plan"
	OperationReprovisionSandboxPlan      SignalType = "reprovision-sandbox-plan"
	OperationReprovisionSandboxApplyPlan SignalType = "reprovision-sandbox-apply-plan"

	OperationReprovisionRunner SignalType = "reprovision-runner"
)

// AllSignalTypes returns all known signal type strings.
func AllSignalTypes() []string {
	return []string{
		OperationForget,
		OperationRestart,
		OperationRestartChildren,
		OperationSyncActionWorkflowTriggers,
		OperationActionWorkflowRun,
		OperationPollDependencies,
		OperationCreated,
		OperationUpdated,
		OperationGenerateInstallStackVersion,
		OperationProvisionRunner,
		OperationAwaitInstallStackVersionRun,
		OperationUpdateInstallStackOutputs,
		OperationAwaitRunnerHealthy,
		OperationProvisionSandbox,
		OperationProvisionDNS,
		OperationDeprovisionDNS,
		OperationExecuteActionWorkflow,
		OperationExecuteDeployComponent,
		OperationExecuteTeardownComponent,
		OperationSyncSecrets,
		OperationWorkflowApproveAll,
		OperationGenerateState,
		OperationExecuteFlow,
		OperationRerunFlow,
		OperationExecuteDeployComponentSyncImage,
		OperationExecuteDeployComponentSyncAndPlan,
		OperationExecuteDeployComponentApplyPlan,
		OperationExecuteDeployComponentPlanOnly,
		OperationExecuteTeardownComponentSyncAndPlan,
		OperationExecuteTeardownComponentApplyPlan,
		OperationProvisionSandboxPlan,
		OperationProvisionSandboxApplyPlan,
		OperationDeprovisionSandboxPlan,
		OperationDeprovisionSandboxApplyPlan,
		OperationReprovisionSandboxPlan,
		OperationReprovisionSandboxApplyPlan,
		OperationReprovisionRunner,
	}
}

type InstallActionWorkflowTriggerSubSignal struct {
	InstallActionWorkflowID string                        `json:"install-action-workflow-id"`
	TriggerType             app.ActionWorkflowTriggerType `json:"trigger-type"`
	TriggeredByID           string                        `json:"triggered-by-id"`
	TriggeredByType         string                        `json:"triggered-by-type"`
	RunEnvVars              map[string]string             `json:"run-env-vars"`
	Role                    string                        `json:"role,omitempty"`
}

type DeployComponentSubSignal struct {
	DeployID    string
	ComponentID string
	CreatePlan  bool
	PlanID      string
	Role        string
}

type TeardownComponentSubSignal struct {
	ComponentID string
	CreatePlan  bool
	PlanID      string
	Role        string
}

type SandboxSubSignal struct {
	CreatePlan     bool
	SkipSyncStatus bool
	PlanID         string
	Role           string
}

type SkipStepSubSignal struct {
	Reason string
}

type RerunOperation string

const (
	RerunOperationSkipStep  RerunOperation = "skip-step"
	RerunOperationRetryStep RerunOperation = "retry-step"
)

type RerunConfiguration struct {
	StepID        string         `json:"step_id"`
	StepOperation RerunOperation `json:"step_operation"`
	StalePlan     bool           `json:"stale_plan"`
	RePlanStepID  string         `json:"replan_step_id"`
}

// Signal contains the details of an install signal operation.
type Signal struct {
	Type SignalType `json:"type"`

	DeployID            string `validate:"required_if=Operation deploy" json:"deploy_id"`
	ActionWorkflowRunID string `validate:"required_if=Operation action_workflow_run" json:"action_workflow_run_id"`
	ForceDelete         bool   `json:"force_delete"`
	InstallWorkflowID   string `validate:"required_if=Operation execute_workflow"`
	FlowID              string `validate:"required_if=Operation execute_flow"`

	InstallActionWorkflowTrigger      InstallActionWorkflowTriggerSubSignal `json:"install_action_workflow_trigger"`
	ExecuteDeployComponentSubSignal   DeployComponentSubSignal              `json:"deploy_component_sub_signal"`
	ExecuteTeardownComponentSubSignal TeardownComponentSubSignal            `json:"teardown_component_sub_signal"`
	ExecuteSkipStepSubSignal          SkipStepSubSignal                     `json:"skip_step_sub_signal"`
	SandboxSubSignal                  SandboxSubSignal                      `json:"sandbox_sub_signal"`

	WorkflowStepID   string `json:"install_workflow_step_id"`
	WorkflowStepName string `json:"install_workflow_step_name"`
	FlowStepID       string `json:"flow_step_id"`
	FlowStepName     string `json:"flow_step_name"`

	RerunConfiguration RerunConfiguration `validate:"required_if=Operation rerun_flow" json:"rerun_configuration"`

	InstallCloudFormationStackVersionID string `json:"install_cloud_formation_stack_version_id"`

	InstallStackID        string `json:"install_stack_id"`
	InstallStackVersionID string `json:"install_stack_version_id"`

	SkipInputUpdateWorkflow bool `json:"skip_input_update_workflow"`
}

// WorkflowID returns the standard event loop workflow ID for the given entity ID.
func (s *Signal) WorkflowID(id string) string {
	return "event-loop-" + id
}

// WorkflowName returns the canonical workflow name.
func (s *Signal) WorkflowName() string {
	return "EventLoop"
}

// EventLoopRequest holds the core request fields previously provided by the eventloop package.
type EventLoopRequest struct {
	ID          string
	SandboxMode bool

	Version            string
	RestartCount       int
	VersionChangeCount int
}

// RequestSignal is the parameter type for install workflow functions.
type RequestSignal struct {
	*Signal
	EventLoopRequest
	StartFromStepIdx int
}

// NewRequestSignal constructs a RequestSignal from an EventLoopRequest and a Signal.
func NewRequestSignal(req EventLoopRequest, signal *Signal) RequestSignal {
	return RequestSignal{
		Signal:           signal,
		EventLoopRequest: req,
	}
}
