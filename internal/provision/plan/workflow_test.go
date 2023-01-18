package plan

import (
	"context"
	"testing"

	"github.com/powertoolsdev/go-generics"
	deploymentplanv1 "github.com/powertoolsdev/protos/deployments/generated/types/plan/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/instances/v1/plan/v1"
	workers "github.com/powertoolsdev/workers-instances/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"google.golang.org/protobuf/proto"
)

func TestProvision(t *testing.T) {
	cfg := generics.GetFakeObj[workers.Config]()
	req := generics.GetFakeObj[*planv1.CreatePlanRequest]()
	wkflow := NewWorkflow(cfg)
	testSuite := &testsuite.WorkflowTestSuite{}
	planRef := generics.GetFakeObj[*deploymentplanv1.PlanRef]()

	// register activities
	a := NewActivities()
	env := testSuite.NewTestWorkflowEnvironment()
	env.OnActivity(a.CreatePlan, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r CreatePlanRequest) (CreatePlanResponse, error) {
			var resp CreatePlanResponse
			//assert.Nil(t, r.validate())
			//assert.Equal(t, cfg, r.Config)

			resp.Plan = planRef
			return resp, nil
		})

	// execute workflow
	env.ExecuteWorkflow(wkflow.Plan, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// verify expected workflow response
	resp := &planv1.CreatePlanResponse{}
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
	assert.True(t, proto.Equal(planRef, resp.Plan))
}
