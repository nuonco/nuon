package worker

import (
	"fmt"
	"strings"
	"time"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/generics"
	execv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
)

func (w *Workflows) execCreatePlanWorkflow(
	ctx workflow.Context,
	sandboxMode bool,
	workflowID string,
	req *planv1.CreatePlanRequest,
) (*planv1.CreatePlanResponse, error) {
	l := workflow.GetLogger(ctx)
	if sandboxMode {
		l.Debug("sandbox-mode enabled, sleeping for to mimic executing plan", zap.String("duration", w.cfg.SandboxSleep.String()))
		workflow.Sleep(ctx, w.cfg.SandboxSleep)
		resp := &planv1.CreatePlanResponse{}
		resp.Ref = generics.GetFakeObj[*planv1.PlanRef]()
		return resp, nil
	}

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:                workflows.ExecutorsTaskQueue,
		WorkflowID:               workflowID,
		WorkflowExecutionTimeout: time.Minute * 60,
		WorkflowTaskTimeout:      time.Minute * 30,
		WorkflowIDReusePolicy:    enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		WaitForCancellation:      true,
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
	l := workflow.GetLogger(ctx)
	if dryRun {
		l.Debug("sandbox-mode enabled, sleeping for to mimic executing plan", zap.String("duration", w.cfg.SandboxSleep.String()))
		workflow.Sleep(ctx, w.cfg.SandboxSleep)
		return generics.GetFakeObj[*execv1.ExecutePlanResponse](), nil
	}

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:                workflows.ExecutorsTaskQueue,
		WorkflowID:               workflowID,
		WorkflowExecutionTimeout: time.Minute * 60,
		WorkflowTaskTimeout:      time.Minute * 30,
		WorkflowIDReusePolicy:    enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		WaitForCancellation:      true,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	var resp execv1.ExecutePlanResponse
	fut := workflow.ExecuteChildWorkflow(ctx, "ExecutePlan", req)
	if err := fut.Get(ctx, &resp); err != nil {
		return &resp, fmt.Errorf("unable to get workflow response: %w", err)
	}
	return &resp, nil
}

func (w *Workflows) execDeprovisionWorkflow(
	ctx workflow.Context,
	dryRun bool,
	req *installsv1.DeprovisionRequest,
) (*installsv1.DeprovisionResponse, error) {
	l := workflow.GetLogger(ctx)
	if dryRun {
		l.Debug("sandbox-mode enabled, sleeping for to mimic provisioning", zap.String("duration", w.cfg.SandboxSleep.String()))
		workflow.Sleep(ctx, w.cfg.SandboxSleep)
		return generics.GetFakeObj[*installsv1.DeprovisionResponse](), nil
	}

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:                workflows.DefaultTaskQueue,
		WorkflowID:               fmt.Sprintf("%s-deprovision", req.InstallId),
		WorkflowExecutionTimeout: time.Minute * 60,
		WorkflowTaskTimeout:      time.Minute * 30,
		WorkflowIDReusePolicy:    enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		WaitForCancellation:      true,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	var resp installsv1.DeprovisionResponse
	fut := workflow.ExecuteChildWorkflow(ctx, "Deprovision", req)
	if err := fut.Get(ctx, &resp); err != nil {
		return &resp, fmt.Errorf("unable to get workflow response: %w", err)
	}
	return &resp, nil
}

func (w *Workflows) execProvisionWorkflow(
	ctx workflow.Context,
	sandboxMode bool,
	req *installsv1.ProvisionRequest,
) (*installsv1.ProvisionResponse, error) {
	l := workflow.GetLogger(ctx)
	if sandboxMode {
		l.Debug("sandbox-mode enabled, sleeping for to mimic provisioning", zap.String("duration", w.cfg.SandboxSleep.String()))
		workflow.Sleep(ctx, w.cfg.SandboxSleep)
		return generics.GetFakeObj[*installsv1.ProvisionResponse](), nil
	}

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:                workflows.DefaultTaskQueue,
		WorkflowID:               fmt.Sprintf("%s-provision", req.InstallId),
		WorkflowExecutionTimeout: time.Minute * 60,
		WorkflowTaskTimeout:      time.Minute * 30,
		WorkflowIDReusePolicy:    enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		WaitForCancellation:      true,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	var resp installsv1.ProvisionResponse
	fut := workflow.ExecuteChildWorkflow(ctx, "Provision", req)
	if err := fut.Get(ctx, &resp); err != nil {
		return &resp, fmt.Errorf("unable to get workflow response: %w", err)
	}

	return &resp, nil
}

func (w *Workflows) execChildWorkflow(
	ctx workflow.Context,
	installID string,
	workflowName string,
	sandboxMode bool,
	req interface{},
	resp interface{},
) error {
	if sandboxMode {
		l := workflow.GetLogger(ctx)
		l.Debug("sandbox-mode enabled, sleeping for to mimic provisioning", zap.String("duration", w.cfg.SandboxSleep.String()))
		workflow.Sleep(ctx, w.cfg.SandboxSleep)
		return nil
	}

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:                workflows.ExecutorsTaskQueue,
		WorkflowID:               fmt.Sprintf("%s-%s", installID, strings.ToLower(workflowName)),
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
