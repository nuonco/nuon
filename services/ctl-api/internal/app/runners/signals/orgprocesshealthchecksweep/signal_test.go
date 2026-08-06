package orgprocesshealthchecksweep

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	runneractivities "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
)

func TestValidateRequiresOrgID(t *testing.T) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()

	sig := &Signal{}
	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.Validate(ctx)
	})

	require.True(t, env.IsWorkflowCompleted())
	require.ErrorContains(t, env.GetWorkflowError(), "org_id is required")
}

func TestExecuteSinglePage(t *testing.T) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()

	sig := &Signal{OrgID: "orgtest123"}
	env.OnActivity(new(runneractivities.Activities).BatchProcessHealthchecks,
		mock.Anything,
		mock.MatchedBy(func(req runneractivities.BatchProcessHealthchecksRequest) bool {
			return req.OrgID == "orgtest123" && req.CursorID == ""
		}),
	).Return(&runneractivities.BatchProcessHealthchecksResponse{Done: true, Checked: 5, Active: 5}, nil).Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.Execute(ctx)
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}

func TestExecuteThreadsCursorAcrossPages(t *testing.T) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()

	sig := &Signal{OrgID: "orgtest123"}
	env.OnActivity(new(runneractivities.Activities).BatchProcessHealthchecks,
		mock.Anything,
		mock.MatchedBy(func(req runneractivities.BatchProcessHealthchecksRequest) bool {
			return req.CursorID == ""
		}),
	).Return(&runneractivities.BatchProcessHealthchecksResponse{NextCursorID: "rnp500", Checked: 500}, nil).Once()
	env.OnActivity(new(runneractivities.Activities).BatchProcessHealthchecks,
		mock.Anything,
		mock.MatchedBy(func(req runneractivities.BatchProcessHealthchecksRequest) bool {
			return req.CursorID == "rnp500"
		}),
	).Return(&runneractivities.BatchProcessHealthchecksResponse{Done: true, Checked: 12}, nil).Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.Execute(ctx)
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}

func TestExecuteSurfacesBatchError(t *testing.T) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()

	sig := &Signal{OrgID: "orgtest123"}
	env.OnActivity(new(runneractivities.Activities).BatchProcessHealthchecks, mock.Anything, mock.Anything).
		Return(nil, errors.New("db down"))

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.Execute(ctx)
	})

	require.True(t, env.IsWorkflowCompleted())
	require.ErrorContains(t, env.GetWorkflowError(), "unable to run process healthcheck batch")
}
