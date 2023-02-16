package plan

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	planactivitiesv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1/activities/v1"
	workers "github.com/powertoolsdev/workers-executors/internal"
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

func (w *wkflow) CreatePlan(ctx workflow.Context, req *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error) {
	resp := &planv1.CreatePlanResponse{}
	l := workflow.GetLogger(ctx)
	ctx = configureActivityOptions(ctx)
	act := NewActivities()

	if err := req.Validate(); err != nil {
		return resp, err
	}

	cpResp, err := w.execCreatePlan(ctx, act, req)
	if err != nil {
		return resp, fmt.Errorf("unable to create plan: %w", err)
	}
	resp.Plan = cpResp.Plan

	l.Debug("successfully created plan for build")
	return resp, nil
}

func (w *wkflow) execCreatePlan(
	ctx workflow.Context,
	act *Activities,
	req *planv1.CreatePlanRequest,
) (*planactivitiesv1.CreatePlanResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := &planactivitiesv1.CreatePlanResponse{}

	l.Debug("executing create plan activity", "request", req)

	switch r := req.Input.(type) {
	case *planv1.CreatePlanRequest_Component:
		actReq, err := w.getComponentPlanRequest(req.Type, r.Component)
		if err != nil {
			return resp, fmt.Errorf("unable to create component plan request: %w", err)
		}
		fut := workflow.ExecuteActivity(ctx, act.CreateComponentPlan, actReq)
		if err := fut.Get(ctx, &resp); err != nil {
			return resp, fmt.Errorf("unable to get component plan future: %w", err)
		}
		return resp, nil

	case *planv1.CreatePlanRequest_Sandbox:
		actReq, err := w.sandboxPlanRequest(req.Type, r.Sandbox)
		if err != nil {
			return resp, fmt.Errorf("unable to create sandbox plan request: %w", err)
		}
		fut := workflow.ExecuteActivity(ctx, act.CreateTerraformSandboxPlan, actReq)
		if err := fut.Get(ctx, &resp); err != nil {
			return resp, fmt.Errorf("unable to get sandbox plan future: %w", err)
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unable to execute plan for type: %s", req.Type)
	}
}
