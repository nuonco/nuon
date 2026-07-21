package automationdispatch

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "automation-dispatch"

const maxAttempts int32 = 5

type Signal struct {
	DispatchID      string `json:"dispatch_id" validate:"required"`
	GenerationToken string `json:"generation_token" validate:"required"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(_ workflow.Context) error { return validator.New().Struct(s) }

func (s *Signal) Execute(ctx workflow.Context) error {
	_, err := activities.AwaitDispatchAutomationEvent(ctx, activities.DispatchAutomationEventRequest{DispatchID: s.DispatchID, GenerationToken: s.GenerationToken}, &workflow.ActivityOptions{
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    15 * time.Second,
			BackoffCoefficient: 2,
			MaximumInterval:    2 * time.Minute,
			MaximumAttempts:    maxAttempts,
		},
	})
	if err != nil {
		if finalizeErr := activities.AwaitFinalizeAutomationDispatchFailure(ctx, activities.FinalizeAutomationDispatchFailureRequest{
			DispatchID:      s.DispatchID,
			GenerationToken: s.GenerationToken,
			Error:           err.Error(),
		}); finalizeErr != nil {
			return errors.Join(err, finalizeErr)
		}
	}
	return err
}
