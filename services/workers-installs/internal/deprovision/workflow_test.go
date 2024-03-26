package deprovision

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	executev1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	workers "github.com/powertoolsdev/mono/services/workers-installs/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
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

func TestDeprovision_finishWithErr(t *testing.T) {
	cfg := generics.GetFakeObj[workers.Config]()
	assert.NoError(t, cfg.Validate())

	req := generics.GetFakeObj[*installsv1.DeprovisionRequest]()
	req.AzureSettings = nil

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	act := NewActivities(nil, nil, nil)

	errChildWorkflow := fmt.Errorf("unable to complete workflow")

	env.RegisterWorkflow(CreatePlan)
	env.OnWorkflow("CreatePlan", mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, pr *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error) {
			return &planv1.CreatePlanResponse{}, errChildWorkflow
		})

	// env.RegisterWorkflow(ExecutePlan)
	// env.OnWorkflow("ExecutePlan", mock.Anything, mock.Anything).
	//	Return(func(_ workflow.Context, pr *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error) {
	//		return &executev1.ExecutePlanResponse{}, errChildWorkflow
	//	})

	env.OnActivity(act.Start, mock.Anything, mock.Anything).
		Return(func(_ context.Context, sReq StartRequest) (StartResponse, error) {
			var resp StartResponse
			assert.NoError(t, sReq.validate())
			return resp, nil
		})

	env.OnActivity(act.FinishDeprovision, mock.Anything, mock.Anything).
		Return(func(_ context.Context, fReq FinishRequest) (FinishResponse, error) {
			var resp FinishResponse
			assert.NoError(t, fReq.validate())

			// verify that when a step fails, the error handler calls finish with the right params
			assert.Contains(t, fReq.ErrorMessage, errChildWorkflow.Error())
			assert.Contains(t, fReq.ErrorStep, "sandbox_plan")
			assert.False(t, fReq.Success)

			return resp, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.Deprovision, req)
}

func TestDeprovision(t *testing.T) {
	cfg := generics.GetFakeObj[workers.Config]()
	assert.NoError(t, cfg.Validate())
	req := generics.GetFakeObj[*installsv1.DeprovisionRequest]()
	req.AzureSettings = nil
	assert.NoError(t, req.Validate())

	planref := generics.GetFakeObj[*planv1.PlanRef]()

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	act := NewActivities(nil, nil, nil)

	env.RegisterWorkflow(CreatePlan)

	env.OnWorkflow("CreatePlan", mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, pr *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error) {
			assert.Nil(t, pr.Validate())

			assert.Equal(t, req.OrgId, pr.GetSandbox().OrgId)
			assert.Equal(t, req.AppId, pr.GetSandbox().AppId)
			assert.Equal(t, req.InstallId, pr.GetSandbox().InstallId)
			assert.Equal(t, req.RunId, pr.GetSandbox().RunId)
			assert.Equal(t, req.SandboxSettings, pr.GetSandbox().SandboxSettings)
			assert.Equal(t, req.AwsSettings, pr.GetSandbox().AwsSettings)

			return &planv1.CreatePlanResponse{Plan: planref}, nil
		})

	env.RegisterWorkflow(ExecutePlan)
	env.OnWorkflow("ExecutePlan", mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, pr *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error) {
			assert.Nil(t, pr.Validate())
			assert.Equal(t, planref, pr.Plan)

			return &executev1.ExecutePlanResponse{}, nil
		})

	// Mock activity implementation
	env.OnActivity(act.Start, mock.Anything, mock.Anything).
		Return(func(_ context.Context, stReq StartRequest) (StartResponse, error) {
			var resp StartResponse
			assert.NoError(t, stReq.validate())

			assert.Equal(t, req.OrgId, stReq.DeprovisionRequest.OrgId)
			assert.Equal(t, req.AppId, stReq.DeprovisionRequest.AppId)
			assert.Equal(t, req.InstallId, stReq.DeprovisionRequest.InstallId)
			return resp, nil
		})

	env.OnActivity(act.FinishDeprovision, mock.Anything, mock.Anything).
		Return(func(_ context.Context, fReq FinishRequest) (FinishResponse, error) {
			var resp FinishResponse
			assert.NoError(t, fReq.validate())

			assert.Equal(t, req.OrgId, fReq.DeprovisionRequest.OrgId)
			assert.Equal(t, req.AppId, fReq.DeprovisionRequest.AppId)
			assert.Equal(t, req.InstallId, fReq.DeprovisionRequest.InstallId)
			return resp, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.Deprovision, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp *installsv1.DeprovisionRequest
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
