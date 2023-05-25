package build

import (
	"context"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	apibuildv1 "github.com/powertoolsdev/mono/pkg/types/api/build/v1"
	executev1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/services/api/internal"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/build/activities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/proto"
)

// NOTE(jm): unfortunately, the only way to register these workflows in the test env is to do it using the same exact
// signature. Given we'll be using these workflows from just about every domain, we should probably make a library to
// wrap these calls, so we don't have to maintain them everywhere like this.
func CreatePlan(workflow.Context, *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error) {
	return nil, nil
}

func ExecutePlan(workflow.Context, *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error) {
	return nil, nil
}

func TestProvision(t *testing.T) {
	cfg := generics.GetFakeObj[*internal.Config]()
	wkflow := New(cfg)
	act := activities.New(nil, "", "")

	// the following fields are used for returning stubbed data
	req := generics.GetFakeObj[*apibuildv1.StartBuildRequest]()
	createPlanReq := &planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Component{
			Component: generics.GetFakeObj[*planv1.ComponentInput](),
		},
	}
	planRef := generics.GetFakeObj[*planv1.PlanRef]()
	buildID := "blduOGobEltmV4wNIYk5AU4t8d"

	// set up our temporal test suite
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(CreatePlan)
	env.RegisterWorkflow(ExecutePlan)
	env.RegisterActivity(act)

	// mock out activities and workflows
	env.OnActivity("CreatePlanRequest", mock.Anything, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *apibuildv1.StartBuildRequest, buildID string) (*planv1.CreatePlanRequest, error) {
			assert.NoError(t, r.Validate())
			return createPlanReq, nil
		})
	env.OnWorkflow("CreatePlan", mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, r *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error) {
			resp := &planv1.CreatePlanResponse{
				Plan: planRef,
			}

			input, ok := r.Input.(*planv1.CreatePlanRequest_Component)
			assert.True(t, ok)
			assert.NotNil(t, input)

			return resp, nil
		})

	env.OnWorkflow("ExecutePlan", mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, r *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error) {
			resp := &executev1.ExecutePlanResponse{}
			assert.True(t, proto.Equal(planRef, r.Plan))
			assert.Nil(t, r.Validate())
			return resp, nil
		})

	// execute workflow
	env.ExecuteWorkflow(wkflow.Build, req, buildID)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// TODO(jm): this needs to use the correct build response
	resp := &planv1.PlanRef{}
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
	assert.True(t, proto.Equal(planRef, resp))
}
