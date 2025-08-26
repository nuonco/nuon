package signals

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

const (
	TemporalNamespace string = "installs"

	OperationForget                      eventloop.SignalType = "forgotten"
	OperationRestart                     eventloop.SignalType = "restart"
	OperationRestartChildren             eventloop.SignalType = "restart-children"
	OperationSyncActionWorkflowTriggers  eventloop.SignalType = "sync-action-workflow-triggers"
	OperationActionWorkflowRun           eventloop.SignalType = "action-workflow-run"
	OperationPollDependencies            eventloop.SignalType = "poll-dependencies"
	OperationCreated                     eventloop.SignalType = "created"
	OperationUpdated                     eventloop.SignalType = "updated"
	OperationGenerateInstallStackVersion eventloop.SignalType = "generate-install-stack-version"
	OperationProvisionRunner             eventloop.SignalType = "provision-runner"
	OperationAwaitInstallStackVersionRun eventloop.SignalType = "await-install-stack-version-run"
	OperationUpdateInstallStackOutputs   eventloop.SignalType = "update-install-stack-outputs"
	OperationAwaitRunnerHealthy          eventloop.SignalType = "await-runner-healthy"
	OperationProvisionSandbox            eventloop.SignalType = "provision-sandbox"
	OperationProvisionDNS                eventloop.SignalType = "provision-dns"
	OperationDeprovisionDNS              eventloop.SignalType = "deprovision-dns"
	OperationExecuteActionWorkflow       eventloop.SignalType = "execute-action-workflow"
	OperationExecuteDeployComponent      eventloop.SignalType = "execute-deploy-component"
	OperationExecuteTeardownComponent    eventloop.SignalType = "execute-teardown-component"
	OperationSyncSecrets                 eventloop.SignalType = "sync-secrets"
	OperationWorkflowApproveAll          eventloop.SignalType = "workflow-approve-all"
	OperationGenerateState               eventloop.SignalType = "generate-state"

	// the following will be sent to a different namespace
	OperationExecuteFlow eventloop.SignalType = "execute-flow"
	OperationRerunFlow   eventloop.SignalType = "rerun-flow"

	// sync images
	OperationExecuteDeployComponentSyncImage eventloop.SignalType = "component-sync-image"

	// approval based deploys
	OperationExecuteDeployComponentSyncAndPlan eventloop.SignalType = "component-deploy-sync-and-plan"
	OperationExecuteDeployComponentApplyPlan   eventloop.SignalType = "component-deploy-apply-plan"
	OperationExecuteDeployComponentPlanOnly    eventloop.SignalType = "component-deploy-plan-only"

	OperationExecuteTeardownComponentSyncAndPlan eventloop.SignalType = "component-teardown-sync-and-plan"
	OperationExecuteTeardownComponentApplyPlan   eventloop.SignalType = "component-teardown-apply-plan"

	// approval based sandbox
	OperationProvisionSandboxPlan        eventloop.SignalType = "provision-sandbox-plan"
	OperationProvisionSandboxApplyPlan   eventloop.SignalType = "provision-sandbox-apply-plan"
	OperationDeprovisionSandboxPlan      eventloop.SignalType = "deprovision-sandbox-plan"
	OperationDeprovisionSandboxApplyPlan eventloop.SignalType = "deprovision-sandbox-apply-plan"
	OperationReprovisionSandboxPlan      eventloop.SignalType = "reprovision-sandbox-plan"
	OperationReprovisionSandboxApplyPlan eventloop.SignalType = "reprovision-sandbox-apply-plan"

	// TODO(jm): deprecate
	OperationReprovisionRunner eventloop.SignalType = "reprovision-runner"
)

type InstallActionWorkflowTriggerSubSignal struct {
	InstallActionWorkflowID string                        `json:"install-action-workflow-id"`
	TriggerType             app.ActionWorkflowTriggerType `json:"trigger-type"`
	TriggeredByID           string                        `json:"triggered-by-id"`
	TriggeredByType         string                        `json:"triggered-by-type"`
	RunEnvVars              map[string]string             `json:"run-env-vars"`
}

type DeployComponentSubSignal struct {
	DeployID    string
	ComponentID string

	// used to control if a plan is created or not
	CreatePlan bool

	// used to consume an existing plan-id
	PlanID string
}

type TeardownComponentSubSignal struct {
	ComponentID string

	// used to control if a plan is created or not
	CreatePlan bool

	// used to consume an existing plan-id
	PlanID string
}

type SandboxSubSignal struct {
	// used to control if a plan is created or not
	CreatePlan bool

	SkipSyncStatus bool

	// used to consume an existing plan-id
	PlanID string
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

type Signal struct {
	Type eventloop.SignalType `json:"type"`

	DeployID            string `validate:"required_if=Operation deploy" json:"deploy_id"`
	ActionWorkflowRunID string `validate:"required_if=Operation action_workflow_run" json:"action_workflow_run_id"`
	ForceDelete         bool   `json:"force_delete"`
	InstallWorkflowID   string `validate:"required_if=Operation execute_workflow"`
	FlowID              string `validate:"required_if=Operation execute_flow"`

	// used for triggering an action workflow
	InstallActionWorkflowTrigger      InstallActionWorkflowTriggerSubSignal `json:"install_action_workflow_trigger"`
	ExecuteDeployComponentSubSignal   DeployComponentSubSignal              `json:"deploy_component_sub_signal"`
	ExecuteTeardownComponentSubSignal TeardownComponentSubSignal            `json:"teardown_component_sub_signal"`
	ExecuteSkipStepSubSignal          SkipStepSubSignal                     `json:"skip_step_sub_signal"`
	SandboxSubSignal                  SandboxSubSignal                      `json:"sandbox_sub_signal"`

	// used for executing an install workflow
	WorkflowStepID   string `json:"install_workflow_step_id"`
	WorkflowStepName string `json:"install_workflow_step_name"`
	FlowStepID       string `json:"flow_step_id"`
	FlowStepName     string `json:"flow_step_name"`

	// used for rerunning an install workflow
	RerunConfiguration RerunConfiguration `validate:"required_if=Operation rerun_flow" json:"rerun_configuration"`

	// used for awaiting the run
	InstallCloudFormationStackVersionID string `json:"install_cloud_formation_stack_version_id"`

	// used for sandbox runs

	eventloop.BaseSignal
}

func NewRequestSignal(req eventloop.EventLoopRequest, signal *Signal) RequestSignal {
	return RequestSignal{
		Signal:           signal,
		EventLoopRequest: req,
	}
}

type RequestSignal struct {
	*Signal `validate:"required"`
	eventloop.EventLoopRequest
}

var _ eventloop.Signal = (*Signal)(nil)

func (s *Signal) ConcurrencyGroup() string {
	switch s.Type {
	case OperationExecuteFlow:
		return "flows"
	default:
		return ""
	}
}

func (s *Signal) Validate(v *validator.Validate) error {
	if err := v.Struct(s); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	return nil
}

func (s *Signal) SignalType() eventloop.SignalType {
	return s.Type
}

func (s *Signal) Namespace() string {
	return TemporalNamespace
}

func (s *Signal) Name() string {
	return string(s.Type)
}

func (s *Signal) Stop() bool {
	switch s.Type {
	default:
	}

	return false
}

func (s *Signal) Restart() bool {
	switch s.Type {
	case OperationRestart:
		return true
	default:
	}

	return false
}

func (s *Signal) Start() bool {
	switch s.Type {
	case OperationCreated:
		return true
	default:
	}

	return false
}

func (s *Signal) GetOrg(ctx context.Context, id string, db *gorm.DB) (*app.Org, error) {
	org, err := cctx.OrgFromContext(ctx)
	if err == nil {
		return org, nil
	}

	install := app.Install{}
	res := db.WithContext(ctx).
		Preload("Org").
		First(&install, "id = ?", id)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return &install.Org, nil
}
