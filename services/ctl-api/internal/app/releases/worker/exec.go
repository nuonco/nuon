package worker

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/workflows"
	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) execProvisionStepWorkflow(
	ctx workflow.Context,
	workflowID string,
	req ProvisionReleaseStepRequest,
) (ProvisionReleaseStepResponse, error) {
	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:                workflows.APITaskQueue,
		WorkflowID:               workflowID,
		WorkflowExecutionTimeout: time.Minute * 60,
		WorkflowTaskTimeout:      time.Minute * 30,
		WorkflowIDReusePolicy:    enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	fut := workflow.ExecuteChildWorkflow(ctx, w.ProvisionReleaseStep, req)
	var resp ProvisionReleaseStepResponse
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, fmt.Errorf("unable to get workflow response: %w", err)
	}

	return resp, nil
}
