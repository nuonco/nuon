package sandbox

import (
	"time"

	executev1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	workflowsclient "github.com/powertoolsdev/mono/pkg/workflows/client"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
)

func Execute(ctx workflow.Context, cpr *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error) {
	resp := &executev1.ExecutePlanResponse{}

	cwo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: time.Minute * 60,
		TaskQueue:                workflowsclient.ExecutorsTaskQueue,
		WorkflowIDReusePolicy:    enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)
	fut := workflow.ExecuteChildWorkflow(ctx, "ExecutePlan", cpr)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}
