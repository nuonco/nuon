package worker

import (
	"fmt"
	"time"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/generics"
	orgsv1 "github.com/nuonco/nuon/pkg/types/workflows/orgs/v1"
	iamv1 "github.com/nuonco/nuon/pkg/types/workflows/orgs/v1/iam/v1"
	"github.com/nuonco/nuon/pkg/workflows"
	orgiam "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/iam"
)

func (w *Workflows) execDeprovisionWorkflow(
	ctx workflow.Context,
	dryRun bool,
	req *orgsv1.DeprovisionRequest,
) (*orgsv1.DeprovisionResponse, error) {
	if dryRun {
		l := workflow.GetLogger(ctx)
		l.Debug("sandbox-mode enabled, sleeping for to mimic deprovisioning", zap.String("duration", w.cfg.SandboxModeSleep.String()))
		workflow.Sleep(ctx, w.cfg.SandboxModeSleep)
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
	req *orgiam.DeprovisionIAMRequest,
) (*iamv1.DeprovisionIAMResponse, error) {
	if dryRun {
		l := workflow.GetLogger(ctx)
		l.Debug("sandbox-mode enabled, sleeping for to mimic provisioning", zap.String("duration", w.cfg.SandboxModeSleep.String()))
		workflow.Sleep(ctx, w.cfg.SandboxModeSleep)
		return generics.GetFakeObj[*iamv1.DeprovisionIAMResponse](), nil
	}

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue:                workflows.DefaultTaskQueue,
		WorkflowID:               fmt.Sprintf("%s-deprovision-iam", req.OrgID),
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
		l.Debug("sandbox-mode enabled, sleeping for to mimic provisioning", zap.String("duration", w.cfg.SandboxModeSleep.String()))
		workflow.Sleep(ctx, w.cfg.SandboxModeSleep)
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
