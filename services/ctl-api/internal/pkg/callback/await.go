package callback

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const defaultAwaitTimeout = 30 * time.Minute

// FallbackAwaitTimeout caps a human-gated wait that has no configured timeout.
const FallbackAwaitTimeout = 30 * 24 * time.Hour

// Result is the payload sent by the handler on completion.
type Result struct {
	Status            string `json:"status"`
	StatusDescription string `json:"status_description,omitempty"`
}

// Await waits for a completion signal on the Ref's signal channel using the
// default timeout.
func Await(ctx workflow.Context, ref Ref) (*Result, error) {
	return AwaitWithTimeout(ctx, ref, defaultAwaitTimeout)
}

// AwaitWithTimeout waits for a completion signal on the Ref's signal channel.
// A timeout <= 0 waits with no wall-clock deadline (for human-gated waits).
func AwaitWithTimeout(ctx workflow.Context, ref Ref, timeout time.Duration) (*Result, error) {
	ch := workflow.GetSignalChannel(ctx, ref.SignalName)

	var result Result
	received := false

	sel := workflow.NewSelector(ctx)
	sel.AddReceive(ch, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &result)
		received = true
	})

	if timeout > 0 {
		timerCtx, timerCancel := workflow.WithCancel(ctx)
		defer timerCancel()
		sel.AddFuture(workflow.NewTimer(timerCtx, timeout), func(f workflow.Future) {})
	}

	sel.Select(ctx)

	if received {
		if result.Status == "error" {
			return nil, temporal.NewNonRetryableApplicationError(
				result.StatusDescription,
				"SIGNAL_FAILED", nil)
		}
		return &result, nil
	}

	return nil, fmt.Errorf("callback timeout: signal not received within %s", timeout)
}
