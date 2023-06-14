package temporal

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/temporal/temporalzap"
	tclient "go.temporal.io/sdk/client"
)

func (t *temporal) ExecuteWorkflowInNamespace(ctx context.Context,
	namespace string,
	options tclient.StartWorkflowOptions,
	workflow interface{},
	args ...interface{}) (tclient.WorkflowRun, error) {
	defaultClient, err := t.getClient()
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	client, err := tclient.NewClientFromExisting(defaultClient, tclient.Options{
		Namespace: namespace,
		Logger:    temporalzap.NewLogger(t.Logger),
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
	defaultClient, err := t.getClient()
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %w", err)
	}

	client, err := tclient.NewClientFromExisting(defaultClient, tclient.Options{
		Namespace: namespace,
		Logger:    temporalzap.NewLogger(t.Logger),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get client in namespace %s: %w", namespace, err)
	}

	return client.GetWorkflow(ctx, workflowID, runID), nil
}

func (t *temporal) CancelWorkflowInNamespace(ctx context.Context,
	namespace string,
	workflowID string,
	runID string) error {
	defaultClient, err := t.getClient()
	if err != nil {
		return fmt.Errorf("unable to get client: %w", err)
	}

	client, err := tclient.NewClientFromExisting(defaultClient, tclient.Options{
		Namespace: namespace,
		Logger:    temporalzap.NewLogger(t.Logger),
	})
	if err != nil {
		return fmt.Errorf("unable to get client in namespace %s: %w", namespace, err)
	}

	if err := client.CancelWorkflow(ctx, workflowID, runID); err != nil {
		return fmt.Errorf("unable to cancel workflow: %w", err)
	}

	return nil
}
