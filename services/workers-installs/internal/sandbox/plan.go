package sandbox

import (
	"time"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	workflowsclient "github.com/powertoolsdev/mono/pkg/workflows/client"
	"go.temporal.io/sdk/workflow"
)

func Plan(ctx workflow.Context, workflowID string, cpr *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error) {
	resp := &planv1.CreatePlanResponse{}

	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:               workflowID,
		WorkflowExecutionTimeout: time.Minute * 5,
		TaskQueue:                workflowsclient.ExecutorsTaskQueue,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)
	fut := workflow.ExecuteChildWorkflow(ctx, "CreatePlan", cpr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
