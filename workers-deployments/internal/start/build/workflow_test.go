package build

import (
	"testing"

	"github.com/powertoolsdev/go-generics"
	buildv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1/build/v1"
	executev1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	workers "github.com/powertoolsdev/workers-deployments/internal"
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
	cfg := generics.GetFakeObj[workers.Config]()
	req := generics.GetFakeObj[*buildv1.BuildRequest]()
	planRef := generics.GetFakeObj[*planv1.PlanRef]()
	wkflow := NewWorkflow(cfg)
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(CreatePlan)
	env.RegisterWorkflow(ExecutePlan)

	env.OnWorkflow("CreatePlan", mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, r *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error) {
			resp := &planv1.CreatePlanResponse{
				Plan: planRef,
			}

			input, ok := r.Input.(*planv1.CreatePlanRequest_Component)
			assert.True(t, ok)

			assert.Equal(t, req.OrgId, input.Component.OrgId)
			assert.Equal(t, req.AppId, input.Component.AppId)
			assert.Equal(t, req.DeploymentId, input.Component.DeploymentId)
			assert.True(t, proto.Equal(req.Component, input.Component.Component))
			assert.Nil(t, r.Validate())
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
	env.ExecuteWorkflow(wkflow.Build, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	resp := &buildv1.BuildResponse{}
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
	assert.True(t, proto.Equal(planRef, resp.PlanRef))
}
