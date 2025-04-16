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
)

// NOTE(jm): we use this pattern in many workflows, and while we do want to move to using signals to make it easier to
// control flow, this approach is still common.
//
// Eventually, this should also allow us to do continue-as-new polling as well.
type PollerFn func(context workflow.Context) error

type PollOpts struct {
	MaxTS           time.Time     `validate:"required"`
	InitialInterval time.Duration `validate:"required"`
	MaxInterval     time.Duration `validate:"required"`
	BackoffFactor   float64       `validate:"required"`
	Fn              PollerFn      `validate:"required"`
}

func Poll(ctx workflow.Context, v *validator.Validate, opts PollOpts) error {
	if err := v.Struct(&opts); err != nil {
		return err
	}

	currentInterval := opts.InitialInterval
	for {
		err := opts.Fn(ctx)
		if err == nil {
			return nil
		}
		if errors.Is(err, NonRetryableError) {
			return err
		}
		if err := workflow.Sleep(ctx, currentInterval); err != nil {
			return errors.Wrap(err, "sleep failed")
		}

		ts := workflow.Now(ctx)
		if ts.After(opts.MaxTS) {
			return context.DeadlineExceeded
		}

		// Increase interval with backoff, but don't exceed MaxInterval
		nextInterval := time.Duration(float64(currentInterval) * opts.BackoffFactor)
		if nextInterval > opts.MaxInterval {
			currentInterval = opts.MaxInterval
		} else {
			currentInterval = nextInterval
		}
	}

	return ExhaustedError
}
