package automationevent

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "automation-event"

type Signal struct {
	EventID  string `json:"event_id" validate:"required"`
	ReplayID string `json:"replay_id,omitempty"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(_ workflow.Context) error {
	return validator.New().Struct(s)
}

func (s *Signal) Execute(ctx workflow.Context) error {
	routed, err := activities.AwaitRouteAutomationEvent(ctx, activities.RouteAutomationEventRequest{EventID: s.EventID, ReplayID: s.ReplayID})
	if err != nil {
		return err
	}
	var dispatchErr error
	for _, dispatch := range routed.Dispatches {
		_, err := activities.AwaitDispatchAutomationEvent(ctx, activities.DispatchAutomationEventRequest{DispatchID: dispatch.ID, GenerationToken: dispatch.GenerationToken}, &workflow.ActivityOptions{
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    15 * time.Second,
				BackoffCoefficient: 2,
				MaximumInterval:    2 * time.Minute,
				MaximumAttempts:    5,
			},
		})
		if err != nil {
			if finalizeErr := activities.AwaitFinalizeAutomationDispatchFailure(ctx, activities.FinalizeAutomationDispatchFailureRequest{
				DispatchID:      dispatch.ID,
				GenerationToken: dispatch.GenerationToken,
				Error:           err.Error(),
			}); finalizeErr != nil {
				err = errors.Join(err, finalizeErr)
			}
		}
		dispatchErr = errors.Join(dispatchErr, err)
	}
	return dispatchErr
}
