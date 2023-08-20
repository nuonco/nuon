package temporal

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/temporal/temporalzap"
	tclient "go.temporal.io/sdk/client"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_client.go -source=client.go -package=temporal
type Client interface {
	tclient.Client

	GetNamespaceClient(namespace string) (tclient.Client, error)

	// ExecuteWorkflowInNamespace is a wrapper that will execute a workflow in a different namespace
	ExecuteWorkflowInNamespace(ctx context.Context,
		namespace string,
		options tclient.StartWorkflowOptions,
		workflow interface{},
		args ...interface{}) (tclient.WorkflowRun, error)

	// ExecuteWorkflowInNamespace is a wrapper that will get a workflow in a different namespace
	GetWorkflowInNamespace(ctx context.Context,
		namespace string,
		workflowID string,
		runID string) (tclient.WorkflowRun, error)

	// CancelWorkflowInNamespace is a wrapper that will get a workflow in a different namespace
	CancelWorkflowInNamespace(ctx context.Context,
		namespace string,
		workflowID string,
		runID string) error

	// SignalWorkflowInNamespace is a wrapper that will signal a workflow in a different namespace
	SignalWorkflowInNamespace(ctx context.Context,
		namespace string,
		workflowID string,
		runID string,
		signalName string,
		signalArg interface{}) error

	// SignalWithStartWorkflowInNamespace is a wrapper that will signal and start a workflow in a different
	// namespace
	SignalWithStartWorkflowInNamespace(ctx context.Context,
		namespace string,
		workflowID string,
		signalName string,
		signalArg interface{},
		options tclient.StartWorkflowOptions,
		workflow interface{},
		workflowArgs interface{}) (tclient.WorkflowRun, error)
}

// getClient returns a temporal client from memory, or creates a new one and caches it
func (t *temporal) getClient() (tclient.Client, error) {
	t.RLock()
	client := t.Client
	t.RUnlock()
	if client != nil {
		return client, nil
	}

	// no client was found, create a new one, set it and return it
	tc, err := tclient.Dial(tclient.Options{
		HostPort:  t.Addr,
		Namespace: t.Namespace,
		Logger:    temporalzap.NewLogger(t.Logger),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to dial temporal: %w", err)
	}

	t.Lock()
	defer t.Unlock()
	t.Client = tc
	return tc, nil
}
