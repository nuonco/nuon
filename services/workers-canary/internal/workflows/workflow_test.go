package workflows

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	activitiesv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1/activities/v1"
	workers "github.com/powertoolsdev/mono/services/workers-canary/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestProvision(t *testing.T) {
	// TODO(jm): re-enable this once this workflow is working correctly
	return
	v := validator.New()
	cfg := generics.GetFakeObj[workers.Config]()

	wkflow, err := New(v, cfg)
	assert.NoError(t, err)

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	req := generics.GetFakeObj[*canaryv1.ProvisionRequest]()

	// Mock activity implementations
	env.OnActivity(activitiesv1.Activity_ACTIVITY_POLL_WORKFLOW.String(), mock.Anything, mock.Anything).
		Return(func(ctx context.Context, aReq *activitiesv1.PollWorkflowRequest) (*activitiesv1.PollWorkflowResponse, error) {
			return &activitiesv1.PollWorkflowResponse{}, nil
		})

	env.OnActivity(activitiesv1.Activity_ACTIVITY_START_WORKFLOW.String(), mock.Anything, mock.Anything).
		Return(func(ctx context.Context, aReq *activitiesv1.StartWorkflowRequest) (*activitiesv1.StartWorkflowResponse, error) {
			return &activitiesv1.StartWorkflowResponse{}, nil
		})

	// execute workflow
	env.ExecuteWorkflow(wkflow.Provision, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// verify response
	var resp *canaryv1.ProvisionResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
