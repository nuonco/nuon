package execute

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	executev1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/execute/v1"
	workers "github.com/powertoolsdev/workers-executors/internal"
)

const (
	defaultActivityTimeout = 45 * time.Minute
)

func configureActivityOptions(ctx workflow.Context) workflow.Context {
	return workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultActivityTimeout,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})
}

type wkflow struct {
	cfg workers.Config
}

func NewWorkflow(cfg workers.Config) *wkflow {
	return &wkflow{
		cfg: cfg,
	}
}

func (w *wkflow) ExecutePlan(ctx workflow.Context, req *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error) {
	resp := &executev1.ExecutePlanResponse{}
	l := workflow.GetLogger(ctx)
	ctx = configureActivityOptions(ctx)
	act := NewActivities()

	if err := req.Validate(); err != nil {
		return resp, err
	}

	epResp, err := execExecutePlanLocally(ctx, act, req)
	if err != nil {
		return resp, fmt.Errorf("unable to execute plan locally: %w", err)
	}
	l.Debug("successfully executed plan")

	resp.Outputs = epResp.Outputs

	return resp, nil
}

func execExecutePlanLocally(
	ctx workflow.Context,
	act *Activities,
	req *executev1.ExecutePlanRequest,
) (*executev1.ExecutePlanResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := &executev1.ExecutePlanResponse{}

	l.Debug("executing plan locally", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.ExecutePlanLocally, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
