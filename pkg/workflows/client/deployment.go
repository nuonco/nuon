package client

import (
	"context"
	"fmt"

	deploymentsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deployments/v1"
	tclient "go.temporal.io/sdk/client"
)

func (w *workflowsClient) TriggerDeploymentStart(ctx context.Context, req *deploymentsv1.StartRequest) (string, error) {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: DefaultTaskQueue,
		Memo: map[string]interface{}{
			"org-id":     req.OrgId,
			"app-id":     req.AppId,
			"started-by": w.Agent,
		},
	}

	wfRun, err := w.TemporalClient.ExecuteWorkflow(ctx, opts, "Start", req)
	if err != nil {
		return "", fmt.Errorf("unable to start deployment: %w", err)
	}

	return wfRun.GetID(), nil
}

func (w *workflowsClient) ExecDeploymentStart(ctx context.Context, req *deploymentsv1.StartRequest) (*deploymentsv1.StartResponse, error) {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: DefaultTaskQueue,
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"org-id":        req.OrgId,
			"app-id":        req.AppId,
			"deployment-id": req.DeploymentId,
			"component-id":  req.Component.Id,
			"started-by":    w.Agent,
		},
	}

	resp := &deploymentsv1.StartResponse{}
	fut, err := w.TemporalClient.ExecuteWorkflow(ctx, opts, "Start", req)
	if err != nil {
		return nil, fmt.Errorf("unable to start deployment: %w", err)
	}

	if err := fut.Get(ctx, resp); err != nil {
		return nil, fmt.Errorf("unable to get response: %w", err)
	}

	return resp, nil
}
