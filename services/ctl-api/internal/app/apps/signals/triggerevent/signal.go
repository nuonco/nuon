package triggerevent

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/triggereventdispatch"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "trigger-event"

type Signal struct {
	EventID                string `json:"event_id" validate:"required"`
	ReplayID               string `json:"replay_id,omitempty"`
	RoutingGenerationToken string `json:"routing_generation_token,omitempty"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(_ workflow.Context) error {
	return validator.New().Struct(s)
}

// Execute routes the event, then fans each dispatch out to its own durable
// trigger-event-dispatch signal on the target app's trigger queue so that a
// failing dispatch or waiter never blocks its siblings. Waiter notifications
// run in parallel with bounded retries for the same reason.
func (s *Signal) Execute(ctx workflow.Context) error {
	routed, err := activities.AwaitRouteTriggerEvent(ctx, activities.RouteTriggerEventRequest{EventID: s.EventID, ReplayID: s.ReplayID, RoutingGenerationToken: s.RoutingGenerationToken})
	if err != nil {
		return err
	}
	boundedRetries := &workflow.ActivityOptions{RetryPolicy: &temporal.RetryPolicy{
		InitialInterval: 5 * time.Second, MaximumInterval: time.Minute, BackoffCoefficient: 2, MaximumAttempts: 5,
	}}
	var dispatchErr error
	for _, dispatch := range routed.Dispatches {
		dedupeKey := "trigger-event-dispatch:enqueue:" + dispatch.ID + ":" + dispatch.GenerationToken
		if _, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   dispatch.AppID,
			OwnerType: "apps",
			QueueName: queue.AppTriggersQueueName,
			Signal: &triggereventdispatch.Signal{
				DispatchID:      dispatch.ID,
				GenerationToken: dispatch.GenerationToken,
			},
			DedupeKey:       &dedupeKey,
			SignalOwnerID:   dispatch.ID,
			SignalOwnerType: "event_dispatches",
		}, boundedRetries); err != nil {
			dispatchErr = errors.Join(dispatchErr, err)
		}
	}
	wg := workflow.NewWaitGroup(ctx)
	for _, waiter := range routed.Waiters {
		waiter := waiter
		wg.Add(1)
		workflow.Go(ctx, func(ctx workflow.Context) {
			defer wg.Done()
			if err := activities.AwaitNotifyEventRunbookWaiter(ctx, activities.NotifyEventRunbookWaiterRequest{WaiterID: waiter.ID, OrgID: waiter.OrgID, QueueSignalID: waiter.QueueSignalID}, boundedRetries); err != nil {
				dispatchErr = errors.Join(dispatchErr, err)
			}
		})
	}
	wg.Wait(ctx)
	return dispatchErr
}
