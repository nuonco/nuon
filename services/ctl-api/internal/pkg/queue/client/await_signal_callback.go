package client

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

const defaultCallbackAwaitTimeout = 30 * time.Minute

// AwaitQueueSignalViaCallback waits for a handler completion signal on a
// pre-registered signal channel. This replaces the AwaitSignal activity which
// blocks an activity worker slot with heartbeats.
//
// The callbackSignalName must match the signal name passed to
// EnqueueSignalToOwner so the handler knows where to send the completion signal.
//
// Returns the handler's FinishedResponse, or an error if the signal failed
// or the callback timed out.
func AwaitQueueSignalViaCallback(ctx workflow.Context, callbackSignalName string, queueSignalID string) (*handler.FinishedResponse, error) {
	ch := workflow.GetSignalChannel(ctx, callbackSignalName)

	var result handler.FinishedResponse
	received := false

	timerCtx, timerCancel := workflow.WithCancel(ctx)
	defer timerCancel()

	sel := workflow.NewSelector(ctx)

	sel.AddReceive(ch, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &result)
		received = true
	})

	sel.AddFuture(workflow.NewTimer(timerCtx, defaultCallbackAwaitTimeout), func(f workflow.Future) {
		// Timeout — will fall back to DB check below.
	})

	sel.Select(ctx)

	if received {
		if result.Status == app.StatusError {
			return nil, temporal.NewNonRetryableApplicationError(
				result.StatusDescription,
				"SIGNAL_FAILED", nil)
		}
		return &result, nil
	}

	// Timeout: fall back to DB status check.
	fresh, err := AwaitGetQueueSignal(ctx, queueSignalID)
	if err != nil {
		return nil, errors.Wrap(err, "callback timeout and unable to check DB status")
	}

	if isTerminalStatus(fresh.Status.Status) {
		return terminalResponse(fresh.Status.Status, fresh.Status.StatusHumanDescription)
	}

	return nil, fmt.Errorf("callback timeout: handler did not complete within %s", defaultCallbackAwaitTimeout)
}

// CallbackSignalName returns a deterministic signal name for a given queue signal ID.
// Used by both the enqueue caller and the handler to agree on the signal channel name.
func CallbackSignalName(queueSignalID string) string {
	return fmt.Sprintf("signal-complete-%s", queueSignalID)
}
