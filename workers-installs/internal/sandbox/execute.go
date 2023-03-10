package sandbox

import (
	"time"

	executev1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/execute/v1"
	"go.temporal.io/sdk/workflow"
)

func Execute(ctx workflow.Context, cpr *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error) {
	resp := &executev1.ExecutePlanResponse{}

	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 60,
		TaskQueue:                "executors",
	}
	ctx = workflow.WithChildOptions(ctx, cwo)
	fut := workflow.ExecuteChildWorkflow(ctx, "ExecutePlan", cpr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
