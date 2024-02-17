package worker

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/pkg/generics"
	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	iamv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/iam/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (w *Workflows) execDeprovisionWorkflow(
	ctx workflow.Context,
	dryRun bool,
	req *orgsv1.DeprovisionRequest,
) (*orgsv1.DeprovisionResponse, error) {
	if dryRun {
		l := workflow.GetLogger(ctx)
		l.Debug("sandbox-mode enabled, sleeping for to mimic deprovisioning", zap.String("duration", w.cfg.SandboxSleep.String()))
		workflow.Sleep(ctx, w.cfg.SandboxSleep)
		return generics.GetFakeObj[*orgsv1.DeprovisionResponse](), nil
	}

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:                workflows.DefaultTaskQueue,
		WorkflowID:               fmt.Sprintf("%s-deprovision", req.OrgId),
		WorkflowExecutionTimeout: time.Minute * 20,
		WorkflowTaskTimeout:      time.Minute * 10,
		WorkflowIDReusePolicy:    enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	var resp orgsv1.DeprovisionResponse
	fut := workflow.ExecuteChildWorkflow(ctx, "Deprovision", req)
	if err := fut.Get(ctx, &resp); err != nil {
		return &resp, fmt.Errorf("unable to get workflow response: %w", err)
	}

	return &resp, nil
}

func (w *Workflows) execDeprovisionIAMWorkflow(
	ctx workflow.Context,
	dryRun bool,
	req *iamv1.DeprovisionIAMRequest,
) (*iamv1.DeprovisionIAMResponse, error) {
	if dryRun {
		l := workflow.GetLogger(ctx)
		l.Debug("sandbox-mode enabled, sleeping for to mimic provisioning", zap.String("duration", w.cfg.SandboxSleep.String()))
		workflow.Sleep(ctx, w.cfg.SandboxSleep)
		return generics.GetFakeObj[*iamv1.DeprovisionIAMResponse](), nil
	}

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:                workflows.DefaultTaskQueue,
		WorkflowID:               fmt.Sprintf("%s-deprovision-iam", req.OrgId),
		WorkflowExecutionTimeout: time.Minute * 1,
		WorkflowTaskTimeout:      time.Minute * 10,
		WorkflowIDReusePolicy:    enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	var resp iamv1.DeprovisionIAMResponse
	fut := workflow.ExecuteChildWorkflow(ctx, "DeprovisionIAM", req)
	if err := fut.Get(ctx, &resp); err != nil {
		return &resp, fmt.Errorf("unable to get workflow response: %w", err)
	}

	return &resp, nil
}

func (w *Workflows) execProvisionWorkflow(
	ctx workflow.Context,
	dryRun bool,
	req *orgsv1.ProvisionRequest,
) (*orgsv1.ProvisionResponse, error) {
	if dryRun {
		l := workflow.GetLogger(ctx)
		l.Debug("sandbox-mode enabled, sleeping for to mimic provisioning", zap.String("duration", w.cfg.SandboxSleep.String()))
		workflow.Sleep(ctx, w.cfg.SandboxSleep)
		return generics.GetFakeObj[*orgsv1.ProvisionResponse](), nil
	}

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:                workflows.DefaultTaskQueue,
		WorkflowID:               fmt.Sprintf("%s-provision", req.OrgId),
		WorkflowExecutionTimeout: time.Minute * 20,
		WorkflowTaskTimeout:      time.Minute * 10,
		WorkflowIDReusePolicy:    enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	var resp orgsv1.ProvisionResponse
	fut := workflow.ExecuteChildWorkflow(ctx, "Provision", req)
	if err := fut.Get(ctx, &resp); err != nil {
		return &resp, fmt.Errorf("unable to get workflow response: %w", err)
	}

	return &resp, nil
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
	if err := fut.Get(ctx, &resp); err != nil {
		return fmt.Errorf("unable to get activity response: %w", err)
	}
	return nil
}
