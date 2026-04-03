package activities

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"
)

func (a *Activities) GetAPIVersion(ctx context.Context) (string, error) {
	return a.cfg.Version, nil
}

func AwaitGetAPIVersion(ctx workflow.Context, opts ...*workflow.ActivityOptions) (string, error) {
	var result string
	options := workflow.GetActivityOptions(ctx)
	options.StartToCloseTimeout = time.Duration(60000000000)

	for _, opt := range opts {
		if opt != nil {
			if opt.StartToCloseTimeout != 0 {
				options.StartToCloseTimeout = opt.StartToCloseTimeout
			}
			if opt.RetryPolicy != nil {
				options.RetryPolicy = opt.RetryPolicy
			}
		}
	}

	ctx = workflow.WithActivityOptions(ctx, options)
	err := workflow.ExecuteActivity(ctx, (*Activities).GetAPIVersion).Get(ctx, &result)
	return result, err
}
