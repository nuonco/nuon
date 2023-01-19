package execute

import (
	"context"
	"testing"

	"github.com/powertoolsdev/go-generics"
	executev1 "github.com/powertoolsdev/protos/workflows/generated/types/instances/v1/execute/v1"
	workers "github.com/powertoolsdev/workers-instances/internal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestProvision(t *testing.T) {
	cfg := generics.GetFakeObj[workers.Config]()
	req := generics.GetFakeObj[*executev1.ExecuteRequest]()
	wkflow := NewWorkflow(cfg)
	testSuite := &testsuite.WorkflowTestSuite{}

	// register activities
	a := NewActivities()
	env := testSuite.NewTestWorkflowEnvironment()
	env.OnActivity(a.ExecutePlanAct, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r ExecutePlanRequest) (ExecutePlanResponse, error) {
			var resp ExecutePlanResponse
			//assert.Nil(t, r.validate())
			//assert.Equal(t, cfg, r.Config)

			return resp, nil
		})

	// execute workflow
	env.ExecuteWorkflow(wkflow.ExecutePlan, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// verify expected workflow response
	resp := &executev1.ExecuteResponse{}
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
