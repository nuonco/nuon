package retry

import (
	"context"
	"time"

	smithytime "github.com/aws/smithy-go/time"
	"github.com/pkg/errors"
)

type retryer struct {
	maxAttempts int
	sleep       time.Duration

	fn          RetryFn
	retryCBHook RetryCBHook

	timeout time.Duration
}

func (r *retryer) Wrap(fn func(context.Context) error) func(context.Context) error {
	r.fn = fn
	return r.exec
}

func (r *retryer) exec(ctx context.Context) error {
	attempt := 0

	if r.timeout > time.Duration(0) {
		var cancelFn context.CancelFunc
		ctx, cancelFn = context.WithTimeout(ctx, r.timeout)
		defer cancelFn()
	}

	var err error
	for attempt <= r.maxAttempts || r.maxAttempts < 0 {
		if attempt > 0 {
			if err = r.retryCBHook(ctx, attempt); err != nil {
				return errors.Wrap(err, "retry callback hook failed")
			}

			smithytime.SleepWithContext(ctx, r.sleep)
		}
		if err = ctx.Err(); err != nil {
			return errors.Wrap(err, "parent context closed")
		}

		if err = r.fn(ctx); err == nil {
			return nil
		}

		if IsNonRetryable(err) {
			return errors.Wrap(err, "not retryable")
		}

		attempt += 1
	}

	return errors.Wrap(err, "maximum attempts reached")
}
