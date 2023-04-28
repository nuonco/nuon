package provision

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/generics"
	executev1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	instancesv1 "github.com/powertoolsdev/mono/pkg/types/workflows/instances/v1"
	sharedv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1"
	workers "github.com/powertoolsdev/mono/services/workers-instances/internal"
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
	wf := NewWorkflow(cfg)
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	req := generics.GetFakeObj[*instancesv1.ProvisionRequest]()
	planRef := generics.GetFakeObj[*planv1.PlanRef]()

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
}
