package temporal

import (
	"context"
	"fmt"

	deploymentsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deployments/v1"
	tclient "go.temporal.io/sdk/client"
)

func (r *repo) TriggerDeploymentStart(ctx context.Context, req *deploymentsv1.StartRequest) error {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: "deployment",
	}

	_, err := r.Client.ExecuteWorkflow(ctx, opts, "Start", req)
	if err != nil {
		return fmt.Errorf("unable to start deployment: %w", err)
	}

	return nil
}

func (r *repo) ExecDeploymentStart(ctx context.Context, req *deploymentsv1.StartRequest) (*deploymentsv1.StartResponse, error) {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: "deployment",
	}

	resp := &deploymentsv1.StartResponse{}
	fut, err := r.Client.ExecuteWorkflow(ctx, opts, "Start", req)
	if err != nil {
		return nil, fmt.Errorf("unable to start deployment: %w", err)
	}

	if err := fut.Get(ctx, resp); err != nil {
		return nil, fmt.Errorf("unable to get response: %w", err)
	}

	return resp, nil
}
