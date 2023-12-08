package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultTerraformRunTimeout time.Duration = time.Minute * 45
)

func (w *wkflow) defaultExecGetActivity(
	ctx workflow.Context,
	actFn interface{},
	req interface{},
	resp interface{},
) error {
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	fut := workflow.ExecuteActivity(ctx, actFn, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return fmt.Errorf("unable to get activity response: %w", err)
	}
	return nil
}

func (w *wkflow) defaultTerraformRunActivity(
	ctx workflow.Context,
	actFn interface{},
	req interface{},
	resp interface{},
	maxAttempts int32,
) error {
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: time.Duration(maxAttempts) * defaultTerraformRunTimeout,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: maxAttempts,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	fut := workflow.ExecuteActivity(ctx, actFn, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return fmt.Errorf("unable to get activity response: %w", err)
	}
	return nil
}

func (w *wkflow) defaultExecTestActivity(
	ctx workflow.Context,
	actFn interface{},
	req interface{},
	resp interface{},
) error {
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	fut := workflow.ExecuteActivity(ctx, actFn, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return fmt.Errorf("unable to get test response: %w", err)
	}
	return nil
}
