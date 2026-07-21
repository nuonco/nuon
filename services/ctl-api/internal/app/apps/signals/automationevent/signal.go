package automationevent

import (
	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "automation-event"

type Signal struct {
	EventID string `json:"event_id" validate:"required"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(_ workflow.Context) error {
	return validator.New().Struct(s)
}

func (s *Signal) Execute(ctx workflow.Context) error {
	_, err := activities.AwaitRouteAutomationEvent(ctx, activities.RouteAutomationEventRequest{EventID: s.EventID})
	return err
}
