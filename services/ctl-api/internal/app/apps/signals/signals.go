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
	TemporalNamespace string = "apps"
)

const (
	OperationCreated          eventloop.SignalType = "created"
	OperationDeleted          eventloop.SignalType = "deleted"
	OperationRestart          eventloop.SignalType = "restart"
	OperationProvision        eventloop.SignalType = "provision"
	OperationPollDependencies eventloop.SignalType = "poll_dependencies"
	OperationDeprovision      eventloop.SignalType = "deprovision"
	OperationReprovision      eventloop.SignalType = "reprovision"
	OperationUpdateSandbox    eventloop.SignalType = "update_sandbox"
	OperationExecuteFlow      eventloop.SignalType = "execute-flow"
)

type Signal struct {
	Type eventloop.SignalType `validate:"required"`

	FlowID string `validate:"required_if=Operation execute_flow"`

	// required for new app config
	AppConfigID string `validate:"required_if=Operation config_created"`

	// required for app sandbox config being updated
	AppSandboxConfigID string `validate:"required_if=Operation sandbox_update"`

	eventloop.BaseSignal
}

func NewRequestSignal(req eventloop.EventLoopRequest, signal *Signal) RequestSignal {
	return RequestSignal{
		Signal:           signal,
		EventLoopRequest: req,
	}
}

type RequestSignal struct {
	*Signal
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
	case OperationDeleted, OperationDeprovision:
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

	currentApp := app.App{}
	res := db.WithContext(ctx).
		Preload("Org").
		First(&currentApp, "id = ?", id)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return currentApp.Org, nil
}
