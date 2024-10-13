package retry

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

const (
	defaultMaxAttempts int           = 5
	defaultSleep       time.Duration = time.Second
)

type retryOpt func(*retryer) error

type RetryFn func(context.Context) error

// retry cb is used to expose a callback hook to users, useful for printing output and more.
type RetryCBHook func(int) error

func noopRetryCBHook(_ int) error { return nil }

func Retry(ctx context.Context, fn RetryFn, opts ...retryOpt) error {
	retryer := &retryer{
		retryCBHook: noopRetryCBHook,
		maxAttempts: defaultMaxAttempts,
		sleep:       defaultSleep,
		fn:          fn,
		timeout:     time.Duration(0),
	}

	for _, opt := range opts {
		if err := opt(retryer); err != nil {
			return errors.Wrap(err, "unable to apply option")
		}
	}

	return retryer.exec(ctx)
}

func WithMaxAttempts(maxAttempts int) retryOpt {
	return func(r *retryer) error {
		r.maxAttempts = maxAttempts
		return nil
	}
}

func WithCBHook(cb RetryCBHook) retryOpt {
	return func(r *retryer) error {
		r.retryCBHook = cb
		return nil
	}
}

func WithSleep(sleep time.Duration) retryOpt {
	return func(r *retryer) error {
		r.sleep = sleep
		return nil
	}
}

func WithTimeout(val time.Duration) retryOpt {
	return func(r *retryer) error {
		r.timeout = val
		return nil
	}
}
