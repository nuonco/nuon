package workflows

import (
	"context"

	tclient "go.temporal.io/sdk/client"
)

type temporalClient interface {
	ExecuteWorkflow(context.Context, tclient.StartWorkflowOptions, interface{}, ...interface{}) (tclient.WorkflowRun, error)
}
