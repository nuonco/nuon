package temporal

import (
	"context"

	tclient "go.temporal.io/sdk/client"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_client_test.go -source=client.go -package=temporal
type temporalClient interface {
	ExecuteWorkflow(context.Context, tclient.StartWorkflowOptions, interface{}, ...interface{}) (tclient.WorkflowRun, error)
}
