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
	OperationSyncActionWorkflowTriggers  eventloop.SignalType = "sync-action-workflow-triggers"
	OperationActionWorkflowRun           eventloop.SignalType = "action-workflow-run"
	OperationPollDependencies            eventloop.SignalType = "poll-dependencies"
	OperationCreated                     eventloop.SignalType = "created"
	OperationGenerateInstallStackVersion eventloop.SignalType = "generate-install-stack-version"
	OperationProvisionRunner             eventloop.SignalType = "provision-runner"
	OperationAwaitInstallStackVersionRun eventloop.SignalType = "await-install-stack-version-run"
	OperationUpdateInstallStackOutputs   eventloop.SignalType = "update-install-stack-outputs"
	OperationAwaitRunnerHealthy          eventloop.SignalType = "await-runner-healthy"
	OperationProvisionSandbox            eventloop.SignalType = "provision-sandbox"
	OperationProvisionDNS                eventloop.SignalType = "provision-dns"
	OperationDeprovisionDNS              eventloop.SignalType = "deprovision-dns"
	OperationDeprovisionSandbox          eventloop.SignalType = "deprovision-sandbox"
	OperationReprovisionSandbox          eventloop.SignalType = "reprovision-sandbox"
	OperationExecuteActionWorkflow       eventloop.SignalType = "trigger-install-action-workflow"
	OperationExecuteDeployComponent      eventloop.SignalType = "execute-deploy-component"
	OperationExecuteTeardownComponent    eventloop.SignalType = "execute-teardown-component"

	// the following will be sent to a different namespace
	OperationExecuteWorkflow eventloop.SignalType = "execute-workflow"

	// the following signals will be deprecated with workflows
	OperationDeploy             eventloop.SignalType = "deploy"
	OperationDeployComponents   eventloop.SignalType = "deploy-components"
	OperationDeleteComponents   eventloop.SignalType = "delete-components"
	OperationTeardownComponents eventloop.SignalType = "teardown-components"
	OperationProvision          eventloop.SignalType = "provision"
	OperationDeprovision        eventloop.SignalType = "deprovision"
	OperationReprovision        eventloop.SignalType = "reprovision"
	OperationReprovisionRunner  eventloop.SignalType = "reprovision-runner"
	OperationDeprovisionRunner  eventloop.SignalType = "deprovision-runner"
	OperationDelete             eventloop.SignalType = "delete"
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
}

type TeardownComponentSubSignal struct {
	ComponentID string
}

type Signal struct {
	Type eventloop.SignalType `json:"type"`

	DeployID            string `validate:"required_if=Operation deploy" json:"deploy_id"`
	ActionWorkflowRunID string `validate:"required_if=Operation action_workflow_run" json:"action_workflow_run_id"`
	ForceDelete         bool   `json:"force_delete"`
	InstallWorkflowID   string `validate:"required_if=Operation execute_workflow"`

	// used for triggering an action workflow
	InstallActionWorkflowTrigger      InstallActionWorkflowTriggerSubSignal `json:"install_action_workflow_trigger"`
	ExecuteDeployComponentSubSignal   DeployComponentSubSignal              `json:"deploy_component_sub_signal"`
	ExecuteTeardownComponentSubSignal TeardownComponentSubSignal            `json:"teardown_component_sub_signal"`

	// used for executing an install workflow
	WorkflowStepID   string `json:"install_workflow_step_id"`
	WorkflowStepName string `json:"install_workflow_step_name"`

	// used for awaiting the run
	InstallCloudFormationStackVersionID string `json:"install_cloud_formation_stack_version_id"`

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
