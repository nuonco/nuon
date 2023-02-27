package execute

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/go-generics"
	executev1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/execute/v1"
	workers "github.com/powertoolsdev/workers-executors/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	env.OnActivity(a.ExecutePlanLocally, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error) {
			resp := &executev1.ExecutePlanResponse{}
			return resp, nil
		})

	// execute workflow
	env.ExecuteWorkflow(wkflow.ExecutePlan, req)
	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())

	// verify expected workflow response
	resp := &executev1.ExecutePlanResponse{}
	assert.NoError(t, env.GetWorkflowResult(&resp))
	assert.NotNil(t, resp)
	env.AssertExpectations(t)
}

func TestExecutePlan_activityError(t *testing.T) {
	cfg := generics.GetFakeObj[workers.Config]()
	req := generics.GetFakeObj[*executev1.ExecutePlanRequest]()
	wkflow := NewWorkflow(cfg)
	testSuite := &testsuite.WorkflowTestSuite{}

	// register activities
	a := NewActivities()
	env := testSuite.NewTestWorkflowEnvironment()
	env.OnActivity(a.ExecutePlanLocally, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error) {
			resp := &executev1.ExecutePlanResponse{}
			return resp, fmt.Errorf("oops")
		})

	// execute workflow
	env.ExecuteWorkflow(wkflow.ExecutePlan, req)
	assert.True(t, env.IsWorkflowCompleted())
	assert.Error(t, env.GetWorkflowError())

	env.AssertExpectations(t)
	// NOTE(jdt): can't use Once or Times on the mocked activity
	env.AssertNumberOfCalls(t, "ExecutePlanLocally", 1)
}
