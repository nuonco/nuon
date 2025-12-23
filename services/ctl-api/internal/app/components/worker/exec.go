package worker

import (
	"fmt"
	"time"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/types/workflows/plan"
	"github.com/nuonco/nuon/pkg/workflows"
)

func (w *Workflows) execCreatePlanWorkflow(
	ctx workflow.Context,
	dryRun bool,
	workflowID string,
	req *plan.CreatePlanRequest,
) (*plan.CreatePlanResponse, error) {
	if dryRun {
		l := workflow.GetLogger(ctx)
		l.Debug("sandbox-mode enabled, sleeping for to mimic executing plan", zap.String("duration", w.cfg.SandboxModeSleep.String()))
		workflow.Sleep(ctx, w.cfg.SandboxModeSleep)
		return plan.FakePlanResponse(), nil
	}

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:                workflows.ExecutorsTaskQueue,
		WorkflowID:               workflowID,
		WorkflowExecutionTimeout: time.Minute * 60,
		WorkflowTaskTimeout:      time.Minute * 30,
		WorkflowIDReusePolicy:    enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	var resp plan.CreatePlanResponse
	fut := workflow.ExecuteChildWorkflow(ctx, "CreatePlan", req)
	if err := fut.Get(ctx, &resp); err != nil {
		return &resp, fmt.Errorf("unable to get workflow response: %w", err)
	}
	return &resp, nil
}
