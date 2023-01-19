package execute

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	executev1 "github.com/powertoolsdev/protos/workflows/generated/types/instances/v1/execute/v1"
	workers "github.com/powertoolsdev/workers-instances/internal"
)

const (
	defaultActivityTimeout = time.Second * 5
)

func configureActivityOptions(ctx workflow.Context) workflow.Context {
	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultActivityTimeout,
	}
	return workflow.WithActivityOptions(ctx, activityOpts)
}

type wkflow struct {
	cfg workers.Config
}

func NewWorkflow(cfg workers.Config) *wkflow {
	return &wkflow{
		cfg: cfg,
	}
}

func (w *wkflow) ExecutePlan(ctx workflow.Context, req *executev1.ExecuteRequest) (*executev1.ExecuteResponse, error) {
	resp := &executev1.ExecuteResponse{}
	l := workflow.GetLogger(ctx)
	ctx = configureActivityOptions(ctx)
	act := NewActivities()

	if err := req.Validate(); err != nil {
		return resp, err
	}

	cpReq := ExecutePlanRequest{}
	_, err := execExecutePlan(ctx, act, cpReq)
	if err != nil {
		return resp, fmt.Errorf("unable to create plan: %w", err)
	}

	l.Debug("successfully created plan for build")
	return resp, nil
}

func execExecutePlan(
	ctx workflow.Context,
	act *Activities,
	req ExecutePlanRequest,
) (ExecutePlanResponse, error) {
	l := workflow.GetLogger(ctx)
	var resp ExecutePlanResponse

	l.Debug("executing execute plan activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.ExecutePlanAct, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
