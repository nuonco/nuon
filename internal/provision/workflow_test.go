package provision

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/go-generics"
	deploymentplanv1 "github.com/powertoolsdev/protos/deployments/generated/types/plan/v1"
	instancesv1 "github.com/powertoolsdev/protos/workflows/generated/types/instances/v1"
	executev1 "github.com/powertoolsdev/protos/workflows/generated/types/instances/v1/execute/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/instances/v1/plan/v1"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
	workers "github.com/powertoolsdev/workers-instances/internal"
	"github.com/powertoolsdev/workers-instances/internal/provision/execute"
	"github.com/powertoolsdev/workers-instances/internal/provision/plan"
)

func TestProvision(t *testing.T) {
	cfg := generics.GetFakeObj[workers.Config]()
	wf := NewWorkflow(cfg)
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	req := generics.GetFakeObj[*instancesv1.ProvisionRequest]()
	planRef := generics.GetFakeObj[*deploymentplanv1.PlanRef]()

	// register child workflows
	pln := plan.NewWorkflow(cfg)
	env.RegisterWorkflow(pln.CreatePlan)
	env.OnWorkflow(pln.CreatePlan, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error) {
			return &planv1.CreatePlanResponse{Plan: planRef}, nil
		})

	exec := execute.NewWorkflow(cfg)
	env.RegisterWorkflow(exec.ExecutePlan)
	env.OnWorkflow(exec.ExecutePlan, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r *executev1.ExecuteRequest) (*executev1.ExecuteResponse, error) {
			return &executev1.ExecuteResponse{}, nil
		})

	// register activities
	a := NewActivities(nil)
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
