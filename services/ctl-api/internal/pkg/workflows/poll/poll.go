package poll

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
)

var (
	NonRetryableError = errors.New("non-retryable")
	ExhaustedError    = errors.New("exhausted attempts new")
	ContinueAsNewErr  = errors.New("continue as new")
)

// NOTE(jm): we use this pattern in many workflows, and while we do want to move to using signals to make it easier to
// control flow, this approach is still common.
//
// Setting ContinueAsNewAfterAttempts to >0 will return ContinueAsNewErr after the given number of attempts.
// Use this to return workflow.ContinueAsNewErr to temporal for doing continue-as-new polling.
type PollerFn func(context workflow.Context) error

type PollOpts struct {
	MaxTS           time.Time     `validate:"required"`
	InitialInterval time.Duration `validate:"required"`
	MaxInterval     time.Duration `validate:"required"`
	BackoffFactor   float64       `validate:"required"`
	// If set to > 0, Poll with return the ContinueAsNewErr after every N attempts.
	ContinueAsNewAfterAttempts int      `validate:"min=0"`
	Fn                         PollerFn `validate:"required"`

	PostAttemptHook func(workflow.Context, time.Duration) error
}

func Poll(ctx workflow.Context, v *validator.Validate, opts PollOpts) error {
	if err := v.Struct(&opts); err != nil {
		return err
	}

	currentIteration := 0
	currentInterval := opts.InitialInterval
	for {
		currentIteration++

		err := opts.Fn(ctx)
		if err == nil {
			return nil
		}
		if errors.Is(err, NonRetryableError) {
			return err
		}

		if opts.PostAttemptHook != nil {
			if err := opts.PostAttemptHook(ctx, currentInterval); err != nil {
				return errors.Wrap(err, "failed in post attempt hook")
			}
		}
		if err := workflow.Sleep(ctx, currentInterval); err != nil {
			return errors.Wrap(err, "sleep failed")
		}

		ts := workflow.Now(ctx)
		if ts.After(opts.MaxTS) {
			return context.DeadlineExceeded
		}

		if opts.ContinueAsNewAfterAttempts > 0 &&
			currentIteration%opts.ContinueAsNewAfterAttempts == 0 {
			return ContinueAsNewErr
		}

		// Increase interval with backoff, but don't exceed MaxInterval
		nextInterval := time.Duration(float64(currentInterval) * opts.BackoffFactor)
		if nextInterval > opts.MaxInterval {
			currentInterval = opts.MaxInterval
		} else {
			currentInterval = nextInterval
		}
	}
}
