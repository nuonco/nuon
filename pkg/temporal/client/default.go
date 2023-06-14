package temporal

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"
)

func (t *temporal) GetWorkflow(ctx context.Context, workflowID string, runID string) tclient.WorkflowRun {
	client, err := t.getClient()
	if err != nil {
		return nil
	}

	return client.GetWorkflow(ctx, workflowID, runID)
}

func (t *temporal) ExecuteWorkflow(ctx context.Context, options tclient.StartWorkflowOptions, workflow interface{}, args ...interface{}) (tclient.WorkflowRun, error) {
	client, err := t.getClient()
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	return client.ExecuteWorkflow(ctx, options, workflow, args...)
}
