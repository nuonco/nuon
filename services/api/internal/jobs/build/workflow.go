package build

import (
	"fmt"
	"time"

	workflowbuildv1 "github.com/powertoolsdev/mono/pkg/types/workflows/builds/v1"
	executev1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/api/internal"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/build/activities"
	"go.temporal.io/sdk/workflow"
)

type wkflow struct {
	cfg *internal.Config
}

func New(cfg *internal.Config) *wkflow {
	return &wkflow{
		cfg: cfg,
	}
}

func (w *wkflow) Build(ctx workflow.Context, req *workflowbuildv1.BuildRequest) (*planv1.PlanRef, error) {
	l := workflow.GetLogger(ctx)

	l.Info("creating plan create request")
	createPlanReq, err := execCreatePlanRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("unable to create plan request: %w", err)
	}
	if err = createPlanReq.Validate(); err != nil {
		return nil, fmt.Errorf("unable to create a plan request: %w", err)
	}

	l.Info("creating plan")
	createPlanResp, err := execCreatePlan(ctx, createPlanReq)
	if err != nil {
		return nil, fmt.Errorf("unable to create plan: %w", err)
	}

	l.Info("executing plan")
	_, err = execExecutePlan(ctx, &executev1.ExecutePlanRequest{
		Plan: createPlanResp.Plan,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create plan: %w", err)
	}

	return createPlanResp.Plan, nil
}

// execCreatePlanRequest returns a plan request that can be passed to a workflow
func execCreatePlanRequest(
	ctx workflow.Context,
	req *workflowbuildv1.BuildRequest,
) (*planv1.CreatePlanRequest, error) {
	resp := &planv1.CreatePlanRequest{}
	l := workflow.GetLogger(ctx)

	// This call with nil is kind of a hacky way to get references to the activity methods,
	// but is not really for code execution since the activity invocations happens
	// over the wire and we can't serialize anything other than pure data arguments
	a := activities.New(nil, "", "")
	opts := workflow.ActivityOptions{ScheduleToCloseTimeout: time.Second * 5}
	ctx = workflow.WithActivityOptions(ctx, opts)

	l.Info("executing create plan request activity")
	fut := workflow.ExecuteActivity(ctx, a.CreatePlanRequest, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
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
		TaskQueue:                workflows.ExecutorsTaskQueue,
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
		TaskQueue:                workflows.ExecutorsTaskQueue,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	fut := workflow.ExecuteChildWorkflow(ctx, "ExecutePlan", req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
