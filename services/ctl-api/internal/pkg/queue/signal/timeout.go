package signal

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DefaultTimeout is the fallback timeout for signals that don't implement
// SignalWithTimeout.
const DefaultTimeout = 30 * 24 * time.Hour

const UnboundedTimeout time.Duration = -1

// DeriveTimeout extracts the timeout from a signal that implements
// SignalWithTimeout. Returns DefaultTimeout if the signal doesn't declare one.
func DeriveTimeout(sig Signal) time.Duration {
	if t, ok := sig.(SignalWithTimeout); ok {
		if t.Timeout() > 0 {
			return t.Timeout()
		}
		if unbounded, ok := sig.(SignalWithUnboundedTimeout); ok && unbounded.UnboundedTimeout() {
			return UnboundedTimeout
		}
	}
	return DefaultTimeout
}

// TimeoutActivityOpts returns an ActivityOptions with ScheduleToCloseTimeout
// set to the given duration. Returns nil when timeout <= 0, which is safe to
// pass to generated Await* wrappers (they skip nil opts).
func TimeoutActivityOpts(timeout time.Duration) *workflow.ActivityOptions {
	if timeout <= 0 {
		return nil
	}
	return &workflow.ActivityOptions{
		ScheduleToCloseTimeout: timeout,
	}
}

// AwaitActivityOpts returns ActivityOptions with the given timeout and a retry
// policy that disables exponential backoff (BackoffCoefficient=1.0), so retries
// happen at a constant interval.
func AwaitActivityOpts(timeout time.Duration) *workflow.ActivityOptions {
	opts := &workflow.ActivityOptions{
		RetryPolicy: &temporal.RetryPolicy{
			BackoffCoefficient: 1.0,
		},
	}
	if timeout > 0 {
		opts.ScheduleToCloseTimeout = timeout
	}
	return opts
}
