package callback

import (
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	// QuickTimeout is for fast infrastructure operations: validation, no-op
	// signals, DB-row creation. If these take longer than 5 minutes something
	// is stuck.
	QuickTimeout = 5 * time.Minute

	// DriftDetectionTimeout is for drift-detection signals which are lightweight
	// but fan out through the webhook/Slack pipeline.
	DriftDetectionTimeout = 15 * time.Minute

	// ShortTimeout is for operations expected to complete within minutes:
	// state generation, lightweight queue signals.
	ShortTimeout = 30 * time.Minute
)

// MaxWaitCeiling bounds every long-lived wait (parked retries, approvals,
// fallback callback waits) so an abandoned workflow closes instead of holding
// its Temporal workflows open indefinitely. A var, not a const, so the flow
// testworker suite can shrink it to exercise expiry paths.
var MaxWaitCeiling = 3 * 24 * time.Hour

// Both snapshot MaxWaitCeiling at init, so overriding it later only affects
// the sites that read MaxWaitCeiling directly.
var (
	// HumanGatedTimeout is for operations that require human interaction:
	// approval workflows, user-initiated stack runs.
	HumanGatedTimeout = MaxWaitCeiling

	// FallbackAwaitTimeout caps a wait that has no configured timeout.
	FallbackAwaitTimeout = MaxWaitCeiling
)

// ErrAwaitTimeout marks an AwaitWithTimeout that expired without receiving
// its completion signal. Match with errors.Is.
var ErrAwaitTimeout = errors.New("callback await timed out")

// Result is the payload sent by the handler on completion.
type Result struct {
	Status            string `json:"status"`
	StatusDescription string `json:"status_description,omitempty"`
}

// CancelledErrType is the application error type returned by AwaitWithTimeout
// when the awaited signal was cancelled. Cancellation must never be treated as
// success — a parent that carries on past a cancelled child silently executes
// steps the user asked to stop.
const CancelledErrType = "SIGNAL_CANCELLED"

// IsCancelled reports whether err (possibly wrapped) is a cancelled-signal
// error from AwaitWithTimeout. Callers use this to stop without invoking
// failure/retry handling.
func IsCancelled(err error) bool {
	var appErr *temporal.ApplicationError
	if errors.As(err, &appErr) {
		return appErr.Type() == CancelledErrType
	}
	return false
}

// cancelledCallbackErrVersion gates cancelled results erroring instead of
// returning as success; in-flight histories carried on past them.
const cancelledCallbackErrVersion = "callback-cancelled-result-err-v1"

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
		// Senders can legitimately arrive with an empty description (status
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
