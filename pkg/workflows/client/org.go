package client

import (
	"context"
	"fmt"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	enumspb "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
)

// TODO(jm): eventually rename this workflow to Provision
func (w *workflowsClient) TriggerOrgSignup(ctx context.Context, req *orgsv1.ProvisionRequest) (string, error) {
	opts := tclient.StartWorkflowOptions{
		ID:                    fmt.Sprintf("%s-provision", req.OrgId),
		TaskQueue:             DefaultTaskQueue,
		WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		Memo: map[string]interface{}{
			"org-id":     req.OrgId,
			"started-by": "nuonctl",
		},
	}

	workflowRun, err := w.TemporalClient.ExecuteWorkflowInNamespace(ctx, "orgs", opts, "Provision", req)
	if err != nil {
		return "", fmt.Errorf("unable to get response: %w", err)
	}

	return workflowRun.GetID(), nil
}

func (w *workflowsClient) TriggerOrgTeardown(ctx context.Context, req *orgsv1.DeprovisionRequest) (string, error) {
	opts := tclient.StartWorkflowOptions{
		ID:                    fmt.Sprintf("%s-deprovision", req.OrgId),
		TaskQueue:             DefaultTaskQueue,
		WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		Memo: map[string]interface{}{
			"org-id":     req.OrgId,
			"started-by": "nuonctl",
		},
	}

	workflowRun, err := w.TemporalClient.ExecuteWorkflowInNamespace(ctx, "orgs", opts, "Deprovision", req)
	if err != nil {
		return "", fmt.Errorf("unable to start teardown: %w", err)
	}

	return workflowRun.GetID(), nil
}

func (w *workflowsClient) ExecOrgSignup(ctx context.Context, req *orgsv1.ProvisionRequest) (*orgsv1.ProvisionResponse, error) {
	opts := tclient.StartWorkflowOptions{
		ID:        fmt.Sprintf("%s-provision", req.OrgId),
		TaskQueue: DefaultTaskQueue,
		Memo: map[string]interface{}{
			"org-id":     req.OrgId,
			"started-by": "nuonctl",
		},
	}

	resp := &orgsv1.ProvisionResponse{}
	fut, err := w.TemporalClient.ExecuteWorkflowInNamespace(ctx, "orgs", opts, "Provision", req)
	if err != nil {
		return nil, fmt.Errorf("unable to start signup: %w", err)
	}

	if err := fut.Get(ctx, resp); err != nil {
		return nil, fmt.Errorf("unable to get response: %w", err)
	}

	return resp, nil
}
