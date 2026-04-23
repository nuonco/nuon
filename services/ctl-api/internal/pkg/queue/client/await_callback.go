package client

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	handleractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler/activities"
)

// AwaitSignalViaCallback enqueues a signal with a callback target pointing at
// the caller workflow, registers a local update handler to receive the
// completion, and blocks until the callback fires.
//
// This is a push-based alternative to AwaitAwaitSignal: no long-running
// heartbeating activity is held while the signal is processed. The caller's
// wait is a pure workflow.Await on a flag set by the update handler.
//
// Constraint: the caller must not continue-as-new between enqueue and
// completion — the update handler registration is lost on CAN, so the
// callback update would fail to deliver.
func AwaitSignalViaCallback(ctx workflow.Context, req *EnqueueSignalRequest) error {
	info := workflow.GetInfo(ctx)
	updateName := "queue_signal_callback_" + req.OwnerID

	var payload handleractivities.CallbackPayload
	var received bool

	err := workflow.SetUpdateHandlerWithOptions(ctx, updateName,
		func(ctx workflow.Context, p handleractivities.CallbackPayload) error {
			payload = p
			received = true
			return nil
		},
		workflow.UpdateHandlerOptions{},
	)
	if err != nil {
		return errors.Wrap(err, "unable to register callback update handler")
	}

	req.Callback = &CallbackRef{
		WorkflowID: info.WorkflowExecution.ID,
		Namespace:  info.Namespace,
		UpdateName: updateName,
	}

	enqueueResp, err := AwaitEnqueueSignal(ctx, req)
	if err != nil {
		return errors.Wrap(err, "unable to enqueue signal")
	}

	// Defensive precheck: if the handler already reached terminal status
	// before the callback landed (unlikely in practice, but cheap insurance),
	// resolve directly from the DB.
	if !received {
		qs, err := activities.AwaitGetQueueSignalByQueueSignalID(ctx, enqueueResp.ID)
		if err == nil && isTerminalStatus(qs.Status.Status) {
			if qs.Status.Status == app.StatusError {
				return temporal.NewNonRetryableApplicationError(
					signalErrorMessage(qs), "SIGNAL_FAILED", nil)
			}
			return nil
		}
	}

	if err := workflow.Await(ctx, func() bool { return received }); err != nil {
		return errors.Wrap(err, "await callback interrupted")
	}

	if !payload.Success {
		return temporal.NewNonRetryableApplicationError(payload.ErrorMessage,
			"SIGNAL_FAILED", nil)
	}
	return nil
}
