package client

import (
	"context"
	"fmt"

	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	enumspb "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

func (w *workflowsClient) TriggerInstallProvision(ctx context.Context, req *installsv1.ProvisionRequest) (string, error) {
	opts := tclient.StartWorkflowOptions{
		ID:        fmt.Sprintf("%s-provision", req.InstallId),
		TaskQueue: DefaultTaskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
		Memo: map[string]interface{}{
			"org-id":     req.OrgId,
			"app-id":     req.AppId,
			"install-id": req.InstallId,
			"started-by": w.Agent,
		},
		WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	workflowRun, err := w.TemporalClient.ExecuteWorkflowInNamespace(ctx, "installs", opts, "Provision", req)
	if err != nil {
		return "", fmt.Errorf("unable to provision install: %w", err)
	}

	return workflowRun.GetID(), nil
}

func (w *workflowsClient) TriggerInstallDeprovision(ctx context.Context, req *installsv1.DeprovisionRequest) (string, error) {
	opts := tclient.StartWorkflowOptions{
		ID:        fmt.Sprintf("%s-deprovision", req.InstallId),
		TaskQueue: DefaultTaskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
		Memo: map[string]interface{}{
			"org-id":     req.OrgId,
			"app-id":     req.AppId,
			"install-id": req.InstallId,
			"started-by": w.Agent,
		},
		WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	workflowRun, err := w.TemporalClient.ExecuteWorkflowInNamespace(ctx, "installs", opts, "Deprovision", req)
	if err != nil {
		return "", fmt.Errorf("unable to deprovision install: %w", err)
	}

	return workflowRun.GetID(), nil
}
