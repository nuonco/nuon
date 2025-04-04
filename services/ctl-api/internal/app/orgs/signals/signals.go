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
	TemporalNamespace string = "orgs"

	OperationCreated         eventloop.SignalType = "created"
	OperationRestart         eventloop.SignalType = "restart"
	OperationRestartRunners  eventloop.SignalType = "restart_runners"
	OperationRestartChildren eventloop.SignalType = "restart_children"

	OperationProvision        eventloop.SignalType = "provision"
	OperationDelete           eventloop.SignalType = "delete"
	OperationForceDelete      eventloop.SignalType = "force_delete"
	OperationDeprovision      eventloop.SignalType = "deprovision"
	OperationForceDeprovision eventloop.SignalType = "force_deprovision"
	OperationReprovision      eventloop.SignalType = "reprovision"
	OperationInviteCreated    eventloop.SignalType = "invite_created"
	OperationInviteAccepted   eventloop.SignalType = "invite_accepted"

	// This signal is only used for stage, when seeding from prod, to ensure an org is set as sandbox mode.
	OperationForceSandboxMode eventloop.SignalType = "force_sandbox_mode"
	OperationStageSeed        eventloop.SignalType = "stage_seed"
)

type Signal struct {
	Type eventloop.SignalType

	// fields for this event loop signal
	InviteID    string
	ForceDelete bool

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
	case OperationForceDelete:
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

	currentOrg := app.Org{}
	res := db.WithContext(ctx).
		First(&currentOrg, "id = ?", id)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org: %w", res.Error)
	}

	return &currentOrg, nil
}

func (s *Signal) Validate(v *validator.Validate) error {
	if err := v.Struct(s); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}
