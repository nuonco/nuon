package temporal

import (
	"context"

	tclient "go.temporal.io/sdk/client"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_client.go -source=client.go -package=temporal
type Client interface {
	tclient.Client

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
}
