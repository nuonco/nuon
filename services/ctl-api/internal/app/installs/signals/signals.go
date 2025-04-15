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

	OperationForgotten                          eventloop.SignalType = "forgotten"
	OperationRestart                            eventloop.SignalType = "restart"
	OperationSyncActionWorkflowTriggers         eventloop.SignalType = "sync_action_workflow_triggers"
	OperationActionWorkflowRun                  eventloop.SignalType = "action_workflow_run"
	OperationPollDependencies                   eventloop.SignalType = "poll_dependencies"
	OperationCreated                            eventloop.SignalType = "created"
	OperationGenerateCloudFormationStackVersion eventloop.SignalType = "generate_cloud_formation_stack_version"
	OperationAwaitCloudFormationStackVersionRun eventloop.SignalType = "await_cloud_formation_stack_version_run"
	OperationAwaitRunnerHealthy                 eventloop.SignalType = "await_runner_healthy"
	OperationProvisionSandbox                   eventloop.SignalType = "provision_sandbox"
	OperationDeprovisionSandbox                 eventloop.SignalType = "deprovision_sandbox"
	OperationReprovisionSandbox                 eventloop.SignalType = "reprovision_sandbox"
	OperationTriggerInstallActionWorkflow       eventloop.SignalType = "trigger_install_action_workflow"
	OperationDeployComponent                    eventloop.SignalType = "deploy_component"
	OperationTeardownComponent                  eventloop.SignalType = "teardown_component"

	// the following will be sent to a different namespace
	OperationExecuteWorkflow eventloop.SignalType = "execute_workflow"

	// the following signals will be deprecated with workflows
	OperationDeploy             eventloop.SignalType = "deploy"
	OperationDeployComponents   eventloop.SignalType = "deploy_components"
	OperationDeleteComponents   eventloop.SignalType = "delete_components"
	OperationTeardownComponents eventloop.SignalType = "teardown_components"
	OperationProvision          eventloop.SignalType = "provision"
	OperationDeprovision        eventloop.SignalType = "deprovision"
	OperationDelete             eventloop.SignalType = "delete"
	OperationReprovision        eventloop.SignalType = "reprovision"
	OperationReprovisionRunner  eventloop.SignalType = "reprovision_runner"
	OperationDeprovisionRunner  eventloop.SignalType = "deprovision_runner"
)

type InstallActionWorkflowTriggerSubSignal struct {
	InstallActionWorkflowID string                        `json:"install_action_workflow_id"`
	TriggerType             app.ActionWorkflowTriggerType `json:"trigger_type"`
	TriggeredByID           string                        `json:"triggered_by_id"`
	TriggeredByType         string                        `json:"triggered_by_type"`
	RunEnvVars              map[string]string             `json:"run_env_vars"`
}

type DeployComponentSubSignal struct {
	DeployID    string
	ComponentID string
}

type TeardownComponentSubSignal struct {
	ComponentID      string
	LatestBuild      bool
	ComponentBuildID string
}

type Signal struct {
        Type eventloop.SignalType `json:"type"`

	DeployID            string `validate:"required_if=Operation deploy" json:"deploy_id"`
	ActionWorkflowRunID string `validate:"required_if=Operation action_workflow_run" json:"action_workflow_run_id"`
	ForceDelete         bool   `json:"force_delete"`
	InstallWorkflowID   string `validate:"required_if=Operation execute_workflow"`

	// used for triggering an action workflow
	InstallActionWorkflowTrigger InstallActionWorkflowTriggerSubSignal `json:"install_action_workflow_trigger"`
	TeardownComponentSubSignal   TeardownComponentSubSignal            `json:"teardown_component_sub_signal"`
	DeployComponentSubSignal     DeployComponentSubSignal              `json:"deploy_component_sub_signal"`

	// used for executing an install workflow
	InstallWorkflowStepID string `json:"install_workflow_step_id"`

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
	case OperationDelete:
		return true
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
