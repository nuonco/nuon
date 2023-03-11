package build

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	buildv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/deployments/v1/build/v1"
	executev1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/executors/v1/plan/v1"
	workers "github.com/powertoolsdev/mono/services/workers-deployments/internal"
)

const (
	defaultExecutorsTaskQueue string = "executors"
)

type wkflow struct {
	cfg workers.Config
}

func NewWorkflow(cfg workers.Config) *wkflow {
	return &wkflow{
		cfg: cfg,
	}
}

func (w *wkflow) Build(ctx workflow.Context, req *buildv1.BuildRequest) (*buildv1.BuildResponse, error) {
	resp := &buildv1.BuildResponse{}
	l := workflow.GetLogger(ctx)

	// TODO(jm): handle noop builds better
	if req.Component.BuildCfg.GetNoop() != nil {
		return resp, nil
	}

	// create plan for container build
	planReq := &planv1.CreatePlanRequest{
		Type: planv1.PlanType_PLAN_TYPE_WAYPOINT_BUILD,
		Input: &planv1.CreatePlanRequest_Component{
			Component: &planv1.Component{
				OrgId:        req.OrgId,
				AppId:        req.AppId,
				DeploymentId: req.DeploymentId,
				Component:    req.Component,
			},
		},
	}
	planResp, err := execCreatePlan(ctx, planReq)
	if err != nil {
		return resp, fmt.Errorf("unable to create plan: %w", err)
	}
	l.Debug("successfully created plan: %v", planResp.Plan)
	resp.PlanRef = planResp.Plan

	if req.PlanOnly {
		l.Debug("plan only enabled, returning")
		return resp, nil
	}

	execReq := &executev1.ExecutePlanRequest{
		Plan: planResp.Plan,
	}
	_, err = execExecutePlan(ctx, execReq)
	if err != nil {
		return resp, fmt.Errorf("unable to execute build plan: %w", err)
	}

	return resp, nil
}

func execCreatePlan(
	ctx workflow.Context,
	req *planv1.CreatePlanRequest,
) (*planv1.CreatePlanResponse, error) {
	resp := &planv1.CreatePlanResponse{}
	l := workflow.GetLogger(ctx)

	l.Debug("executing create plan workflow")
	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 20,
		WorkflowTaskTimeout:      time.Minute * 10,
		TaskQueue:                defaultExecutorsTaskQueue,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	fut := workflow.ExecuteChildWorkflow(ctx, "CreatePlan", req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execExecutePlan(
	ctx workflow.Context,
	req *executev1.ExecutePlanRequest,
) (*executev1.ExecutePlanResponse, error) {
	resp := &executev1.ExecutePlanResponse{}
	l := workflow.GetLogger(ctx)

	l.Debug("executing execute plan workflow")
	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 20,
		WorkflowTaskTimeout:      time.Minute * 10,
		TaskQueue:                defaultExecutorsTaskQueue,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	fut := workflow.ExecuteChildWorkflow(ctx, "ExecutePlan", req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
