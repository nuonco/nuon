package temporal

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"
)

func (t *temporal) ExecuteWorkflowInNamespace(ctx context.Context,
	namespace string,
	options tclient.StartWorkflowOptions,
	workflow interface{},
	args ...interface{}) (tclient.WorkflowRun, error) {
	client, err := tclient.NewClientFromExisting(t.Client, tclient.Options{
		Namespace: namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get client in namespace %s: %w", namespace, err)
	}

	return client.ExecuteWorkflow(ctx, options, workflow, args...)
}

func (t *temporal) GetWorkflowInNamespace(ctx context.Context,
	namespace string,
	workflowID string,
	runID string) (tclient.WorkflowRun, error) {
	client, err := tclient.NewClientFromExisting(t.Client, tclient.Options{
		Namespace: namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get client in namespace %s: %w", namespace, err)
	}

	return client.GetWorkflow(ctx, workflowID, runID), nil
}
