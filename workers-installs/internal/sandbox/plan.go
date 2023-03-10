package sandbox

import (
	"time"

	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"go.temporal.io/sdk/workflow"
)

func Plan(ctx workflow.Context, cpr *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error) {
	resp := &planv1.CreatePlanResponse{}

	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 5,
		TaskQueue:                "executors",
	}
	ctx = workflow.WithChildOptions(ctx, cwo)
	fut := workflow.ExecuteChildWorkflow(ctx, "CreatePlan", cpr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
