package callback

import (
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	QuickTimeout = 5 * time.Minute

	DriftDetectionTimeout = 15 * time.Minute

	ShortTimeout = 30 * time.Minute
)

var MaxWaitCeiling = 3 * 24 * time.Hour

var (
	HumanGatedTimeout = MaxWaitCeiling

	FallbackAwaitTimeout = MaxWaitCeiling
)

var ErrAwaitTimeout = errors.New("callback await timed out")

type Result struct {
	Status            string `json:"status"`
	StatusDescription string `json:"status_description,omitempty"`
}

// why: CancelledErrType is the application error type returned by AwaitWithTimeout
// when the awaited signal was cancelled. Cancellation must never be treated as
// success — a parent that carries on past a cancelled child silently executes
// steps the user asked to stop.
const CancelledErrType = "SIGNAL_CANCELLED"

func IsCancelled(err error) bool {
	var appErr *temporal.ApplicationError
	if errors.As(err, &appErr) {
		return appErr.Type() == CancelledErrType
	}
	return false
}

const cancelledCallbackErrVersion = "callback-cancelled-result-err-v1"

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
		// why: Senders can legitimately arrive with an empty description (status
		// writers that only set Status, cancellations with no error text).
		// The message below becomes user-visible failure text on the parent,
		// so never let it be empty.
		switch result.Status {
		case "error":
			msg := result.StatusDescription
			if msg == "" {
				msg = "failed without further detail"
			}
			return nil, temporal.NewNonRetryableApplicationError(
				msg,
				"SIGNAL_FAILED", nil)
		case "cancelled":
			if workflow.GetVersion(ctx, cancelledCallbackErrVersion, workflow.DefaultVersion, 1) == workflow.DefaultVersion {
				return &result, nil
			}
			msg := result.StatusDescription
			if msg == "" {
				msg = "cancelled"
			}
			return nil, temporal.NewNonRetryableApplicationError(
				msg,
				CancelledErrType, nil)
		}
		return &result, nil
	}

	return nil, fmt.Errorf("callback timeout: signal not received within %s: %w", timeout, ErrAwaitTimeout)
}
