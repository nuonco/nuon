package workflows

import (
	"context"

	"github.com/stretchr/testify/mock"
	tclient "go.temporal.io/sdk/client"
)

type testTemporalClient struct {
	mock.Mock
}

func (t *testTemporalClient) ExecuteWorkflow(ctx context.Context, opts tclient.StartWorkflowOptions, queueName interface{}, args ...interface{}) (tclient.WorkflowRun, error) {
	callArgs := t.Called(ctx, opts, queueName, args)
	if callArgs.Get(0) != nil {
		return callArgs.Get(0).(tclient.WorkflowRun), callArgs.Error(1)
	}

	return nil, callArgs.Error(1)
}

var _ temporalClient = (*testTemporalClient)(nil)
