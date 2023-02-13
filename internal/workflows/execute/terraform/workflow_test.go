package execute

import (
	"context"
	"testing"

	"github.com/powertoolsdev/go-generics"
	executev1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/execute/v1"
	workers "github.com/powertoolsdev/workers-executors/internal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestExecutePlan(t *testing.T) {
	cfg := generics.GetFakeObj[workers.Config]()
	req := generics.GetFakeObj[*executev1.ExecutePlanRequest]()
	wkflow := NewWorkflow(cfg)
	testSuite := &testsuite.WorkflowTestSuite{}

	// register activities
	a := NewActivities()
	env := testSuite.NewTestWorkflowEnvironment()
	env.OnActivity(a.ExecuteTerraformPlanLocally, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error) {
			resp := &executev1.ExecutePlanResponse{}
			return resp, nil
		})

	// execute workflow
	env.ExecuteWorkflow(wkflow.ExecuteTerraformPlan, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// verify expected workflow response
	resp := &executev1.ExecutePlanResponse{}
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
