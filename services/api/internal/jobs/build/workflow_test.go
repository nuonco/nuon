package build

import (
	"context"
	"fmt"
	"testing"

	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/generics"
	apibuildv1 "github.com/powertoolsdev/mono/pkg/types/api/build/v1"
	executev1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/build/activities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

// This test is still WIP
func TestBuildErrorOnCreatePlanRequest(t *testing.T) {

	cfg := generics.GetFakeObj[Config]()
	wf := New(cfg)
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	sbReq := generics.GetFakeObj[*apibuildv1.StartBuildRequest]()
	a := activities.New(nil)
	env.OnActivity(a.CreatePlanRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, sbReq *apibuildv1.StartBuildRequest) (*planv1.CreatePlanRequest, error) {
			return nil, fmt.Errorf("unit test error")
		})

	// exec and assert workflow
	env.ExecuteWorkflow(wf.Build, sbReq, nil)
	assert.True(t, env.IsWorkflowCompleted())
	assert.Error(t, env.GetWorkflowError())
	assert.ErrorContains(t, env.GetWorkflowError(), "unit test error")

	/*
		// register child workflows
		env.RegisterWorkflow(CreatePlan)
		env.OnWorkflow(CreatePlan, mock.Anything, mock.Anything).
			Return(func(ctx workflow.Context, r *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error) {
				assert.NoError(t, r.Validate())
				return &planv1.CreatePlanResponse{Plan: planRef}, nil
			})

		env.RegisterWorkflow(ExecutePlan)
		env.OnWorkflow(ExecutePlan, mock.Anything, mock.Anything).
			Return(func(ctx workflow.Context, r *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error) {
				assert.NoError(t, r.Validate())
				return &executev1.ExecutePlanResponse{}, nil
			})

		// register activities
		a := NewActivities(nil, nil)
		env.OnActivity(a.SendHostnameNotification, mock.Anything, mock.Anything).
			Return(func(_ context.Context, shnReq SendHostnameNotificationRequest) (SendHostnameNotificationResponse, error) {
				var resp SendHostnameNotificationResponse
				assert.Nil(t, shnReq.validate())
				return resp, nil
			})
		env.OnActivity(a.StartProvisionRequest, mock.Anything, mock.Anything).
			Return(func(_ context.Context, r *sharedv1.StartActivityRequest) (*sharedv1.StartActivityResponse, error) {
				return &sharedv1.StartActivityResponse{}, nil
			})
		env.OnActivity(a.FinishProvisionRequest, mock.Anything, mock.Anything).
			Return(func(_ context.Context, r *sharedv1.FinishActivityRequest) (*sharedv1.FinishActivityResponse, error) {
				return &sharedv1.FinishActivityResponse{}, nil
			})

		// exec and assert workflow
		env.ExecuteWorkflow(wf.Provision, req)
		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())

		resp := &instancesv1.ProvisionResponse{}
		require.NoError(t, env.GetWorkflowResult(&resp))
		require.NotNil(t, resp)
	*/
}
