package workflows

import (
	"context"
	"fmt"

	deploymentsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deployments/v1"
	tclient "go.temporal.io/sdk/client"
)

func (wfClient *workflowsClient) TriggerDeploymentStart(ctx context.Context, req *deploymentsv1.StartRequest) error {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: DefaultTaskQueue,
	}

	_, err := wfClient.TemporalClient.ExecuteWorkflow(ctx, opts, "Start", req)
	if err != nil {
		return fmt.Errorf("unable to start deployment: %w", err)
	}

	return nil
}

func (wfClient *workflowsClient) ExecDeploymentStart(ctx context.Context, req *deploymentsv1.StartRequest) (*deploymentsv1.StartResponse, error) {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: DefaultTaskQueue,
	}

	resp := &deploymentsv1.StartResponse{}
	fut, err := wfClient.TemporalClient.ExecuteWorkflow(ctx, opts, "Start", req)
	if err != nil {
		return nil, fmt.Errorf("unable to start deployment: %w", err)
	}

	if err := fut.Get(ctx, resp); err != nil {
		return nil, fmt.Errorf("unable to get response: %w", err)
	}

	return resp, nil
}
