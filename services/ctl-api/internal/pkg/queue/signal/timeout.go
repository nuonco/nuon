package signal

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const DefaultTimeout = 30 * 24 * time.Hour

const UnboundedTimeout time.Duration = -1

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

func TimeoutActivityOpts(timeout time.Duration) *workflow.ActivityOptions {
	if timeout <= 0 {
		return nil
	}
	return &workflow.ActivityOptions{
		ScheduleToCloseTimeout: timeout,
	}
}

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
