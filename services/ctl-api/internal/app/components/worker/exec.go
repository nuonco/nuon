package worker

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) defaultExecGetActivity(
	ctx workflow.Context,
	actFn interface{},
	req interface{},
	resp interface{},
) error {
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 1 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	fut := workflow.ExecuteActivity(ctx, actFn, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return fmt.Errorf("unable to get activity response: %w", err)
	}
	return nil
}

func (w *Workflows) defaultExecErrorActivity(
	ctx workflow.Context,
	actFn interface{},
	req interface{},
) error {
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 1 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	fut := workflow.ExecuteActivity(ctx, actFn, req)
	var respErr error
	if err := fut.Get(ctx, &respErr); err != nil {
		return fmt.Errorf("unable to get activity response: %w", err)
	}

	if respErr != nil {
		return fmt.Errorf("activity returned error: %w", respErr)
	}

	return nil
}
