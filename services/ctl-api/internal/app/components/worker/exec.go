package worker

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/generics"
	execv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (w *Workflows) defaultExecGetActivity(
	ctx workflow.Context,
	actFn interface{},
	req interface{},
	resp interface{},
) error {
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 1 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	fut := workflow.ExecuteActivity(ctx, actFn, req)
	if err := fut.Get(ctx, resp); err != nil {
		return fmt.Errorf("unable to get activity response: %w", err)
	}
	return nil
}

func (w *Workflows) defaultExecErrorActivity(
	ctx workflow.Context,
	actFn interface{},
	req interface{},
) error {
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 1 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	fut := workflow.ExecuteActivity(ctx, actFn, req)
	var respErr error
	if err := fut.Get(ctx, &respErr); err != nil {
		return fmt.Errorf("unable to get activity response: %w", err)
	}

	if respErr != nil {
		return fmt.Errorf("activity returned error: %w", respErr)
	}

	return nil
}

func (w *Workflows) execCreatePlanWorkflow(
	ctx workflow.Context,
	dryRun bool,
	workflowID string,
	req *planv1.CreatePlanRequest,
) (*planv1.CreatePlanResponse, error) {
	if dryRun {
		w.l.Debug("sandbox-mode enabled, sleeping for to mimic executing plan", zap.String("duration", w.cfg.SandboxSleep.String()))
		workflow.Sleep(ctx, w.cfg.SandboxSleep)
		return generics.GetFakeObj[*planv1.CreatePlanResponse](), nil
	}

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:                workflows.ExecutorsTaskQueue,
		WorkflowID:               workflowID,
		WorkflowExecutionTimeout: time.Minute * 60,
		WorkflowTaskTimeout:      time.Minute * 30,
		WorkflowIDReusePolicy:    enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	var resp planv1.CreatePlanResponse
	fut := workflow.ExecuteChildWorkflow(ctx, "CreatePlan", req)
	if err := fut.Get(ctx, &resp); err != nil {
		return &resp, fmt.Errorf("unable to get workflow response: %w", err)
	}
	return &resp, nil
}

func (w *Workflows) execExecPlanWorkflow(
	ctx workflow.Context,
	dryRun bool,
	workflowID string,
	req *execv1.ExecutePlanRequest,
) (*execv1.ExecutePlanResponse, error) {
	if dryRun {
		w.l.Debug("sandbox-mode enabled, sleeping for to mimic executing plan", zap.String("duration", w.cfg.SandboxSleep.String()))
		workflow.Sleep(ctx, w.cfg.SandboxSleep)
		return generics.GetFakeObj[*execv1.ExecutePlanResponse](), nil
	}

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:                workflows.ExecutorsTaskQueue,
		WorkflowID:               workflowID,
		WorkflowExecutionTimeout: time.Minute * 60,
		WorkflowTaskTimeout:      time.Minute * 30,
		WorkflowIDReusePolicy:    enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	var resp execv1.ExecutePlanResponse
	fut := workflow.ExecuteChildWorkflow(ctx, "ExecutePlan", req)
	if err := fut.Get(ctx, &resp); err != nil {
		return &resp, fmt.Errorf("unable to get workflow response: %w", err)
	}
	return &resp, nil
}
