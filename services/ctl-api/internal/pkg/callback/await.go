package callback

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const defaultAwaitTimeout = 5 * time.Minute

// Result is the payload sent by the handler on completion.
type Result struct {
	Status            string `json:"status"`
	StatusDescription string `json:"status_description,omitempty"`
}

// Await waits for a completion signal on the Ref's signal channel.
// Returns the Result on success, or an error if the signal failed or timed out.
func Await(ctx workflow.Context, ref Ref) (*Result, error) {
	ch := workflow.GetSignalChannel(ctx, ref.SignalName)

	var result Result
	received := false

	timerCtx, timerCancel := workflow.WithCancel(ctx)
	defer timerCancel()

	sel := workflow.NewSelector(ctx)

	sel.AddReceive(ch, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &result)
		received = true
	})

	sel.AddFuture(workflow.NewTimer(timerCtx, defaultAwaitTimeout), func(f workflow.Future) {})

	sel.Select(ctx)

	if received {
		if result.Status == "error" {
			return nil, temporal.NewNonRetryableApplicationError(
				result.StatusDescription,
				"SIGNAL_FAILED", nil)
		}
		return &result, nil
	}

	return nil, fmt.Errorf("callback timeout: signal not received within %s", defaultAwaitTimeout)
}
