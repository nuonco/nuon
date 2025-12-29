package worker

import (
	"fmt"
	"strings"
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

func (w *Workflows) execChildWorkflow(
	ctx workflow.Context,
	runnerID string,
	workflowName string,
	sandboxMode bool,
	req interface{},
	resp interface{},
) error {
	if sandboxMode {
		l := workflow.GetLogger(ctx)
		l.Debug("sandbox-mode enabled, sleeping for to mimic provisioning", zap.String("duration", w.cfg.SandboxModeSleep.String()))
		workflow.Sleep(ctx, w.cfg.SandboxModeSleep)
		return nil
	}

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:                workflows.ExecutorsTaskQueue,
		WorkflowID:               fmt.Sprintf("%s-%s", runnerID, strings.ToLower(workflowName)),
		WorkflowExecutionTimeout: time.Minute * 20,
		WorkflowTaskTimeout:      time.Minute * 10,
		WorkflowIDReusePolicy:    enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	fut := workflow.ExecuteChildWorkflow(ctx, workflowName, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return fmt.Errorf("unable to get workflow response: %w", err)
	}

	return nil
}

func (w *Workflows) execHealthcheckChildWorkflow(
	ctx workflow.Context,
	runnerID string,
	workflowName string,
	req interface{},
	resp interface{},
) error {
	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:                workflows.APITaskQueue,
		WorkflowID:               fmt.Sprintf("%s-%s", runnerID, strings.ToLower(workflowName)),
		WorkflowExecutionTimeout: time.Minute * 2,
		WorkflowTaskTimeout:      time.Minute * 10,
		WorkflowIDReusePolicy:    enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	fut := workflow.ExecuteChildWorkflow(ctx, workflowName, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return fmt.Errorf("unable to get workflow response: %w", err)
	}

	return nil
}
