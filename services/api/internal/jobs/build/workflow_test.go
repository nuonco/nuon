package build

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	workflowbuildv1 "github.com/powertoolsdev/mono/pkg/types/workflows/builds/v1"
	executev1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	sharedv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1"
	"github.com/powertoolsdev/mono/pkg/workflows/meta/prefix"
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
	req := generics.GetFakeObj[*workflowbuildv1.BuildRequest]()
	createPlanReq := &planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Component{
			Component: generics.GetFakeObj[*planv1.ComponentInput](),
		},
	}
	planRef := generics.GetFakeObj[*planv1.PlanRef]()

	// set up our temporal test suite
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(CreatePlan)
	env.RegisterWorkflow(ExecutePlan)
	env.RegisterActivity(act)

	// mock out activities and workflows
	env.OnActivity(act.StartStartRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.StartActivityRequest) (*sharedv1.StartActivityResponse, error) {
			resp := &sharedv1.StartActivityResponse{}
			assert.Nil(t, r.Validate())
			assert.Equal(t, cfg.DeploymentsBucket, r.MetadataBucket)

			expectedRoleARN := fmt.Sprintf(cfg.OrgsDeploymentsRoleTemplate, req.OrgId)
			assert.Equal(t, expectedRoleARN, r.MetadataBucketAssumeRoleArn)
			expectedPrefix := prefix.BuildPath(req.OrgId, req.AppId, req.ComponentId, req.BuildId)
			assert.Equal(t, expectedPrefix, r.MetadataBucketPrefix)
			return resp, nil
		})

	env.OnActivity(act.FinishStartRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.FinishActivityRequest) (*sharedv1.FinishActivityResponse, error) {
			resp := &sharedv1.FinishActivityResponse{}
			assert.Nil(t, r.Validate())
			assert.Equal(t, cfg.DeploymentsBucket, r.MetadataBucket)

			expectedRoleARN := fmt.Sprintf(cfg.OrgsDeploymentsRoleTemplate, req.OrgId)
			assert.Equal(t, expectedRoleARN, r.MetadataBucketAssumeRoleArn)
			expectedPrefix := prefix.BuildPath(req.OrgId, req.AppId, req.ComponentId, req.BuildId)
			assert.Equal(t, expectedPrefix, r.MetadataBucketPrefix)
			return resp, nil
		})

	env.OnActivity("CreatePlanRequest", mock.Anything, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *workflowbuildv1.BuildRequest) (*planv1.CreatePlanRequest, error) {
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
	env.ExecuteWorkflow(wkflow.Build, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	resp := &workflowbuildv1.BuildResponse{}
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
	assert.True(t, proto.Equal(planRef, resp.BuildPlan))
}
