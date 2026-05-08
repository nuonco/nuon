package signal

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// DefaultTimeout is the fallback timeout for signals that don't implement
// SignalWithTimeout.
const DefaultTimeout = 2 * time.Hour

// DeriveTimeout extracts the timeout from a signal that implements
// SignalWithTimeout. Returns DefaultTimeout if the signal doesn't declare one.
func DeriveTimeout(sig Signal) time.Duration {
	if t, ok := sig.(SignalWithTimeout); ok && t.Timeout() > 0 {
		return t.Timeout()
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
