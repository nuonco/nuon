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

	OperationCreated                    eventloop.SignalType = "created"
	OperationProvision                  eventloop.SignalType = "provision"
	OperationDeprovision                eventloop.SignalType = "deprovision"
	OperationDelete                     eventloop.SignalType = "delete"
	OperationReprovision                eventloop.SignalType = "reprovision"
	OperationReprovisionRunner          eventloop.SignalType = "reprovision_runner"
	OperationDeprovisionRunner          eventloop.SignalType = "deprovision_runner"
	OperationRestart                    eventloop.SignalType = "restart"
	OperationDeploy                     eventloop.SignalType = "deploy"
	OperationForgotten                  eventloop.SignalType = "forgotten"
	OperationPollDependencies           eventloop.SignalType = "poll_dependencies"
	OperationDeployComponents           eventloop.SignalType = "deploy_components"
	OperationDeleteComponents           eventloop.SignalType = "delete_components"
	OperationActionWorkflowRun          eventloop.SignalType = "action_workflow_run"
	OperationSyncActionWorkflowTriggers eventloop.SignalType = "sync_action_workflow_triggers"

	// DEPRECATED
	// Replaced with OperationDeleteaComponents
	OperationTeardownComponents eventloop.SignalType = "teardown_components"
)

type Signal struct {
	Type eventloop.SignalType

	DeployID            string `validate:"required_if=Operation deploy" json:"deploy_id"`
	ActionWorkflowRunID string `validate:"required_if=Operation action_workflow_run" json:"action_workflow_run_id"`
	ForceDelete         bool   `json:"force_delete"`

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
