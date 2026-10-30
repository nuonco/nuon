package emptygroup

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type emptySig struct {
	empty bool
}

func (s *emptySig) Type() signal.SignalType { return "test-empty-group" }
func (s *emptySig) Validate(ctx workflow.Context) error {
	return nil
}
func (s *emptySig) Execute(ctx workflow.Context) error { return nil }
func (s *emptySig) IsEmptyInstallGroup(ctx workflow.Context) (bool, error) {
	return s.empty, nil
}

var _ signal.Signal = (*emptySig)(nil)
var _ signal.SignalWithEmptyGroupCheck = (*emptySig)(nil)

func TestEmptyGroupCheckCompletesSuccessfully(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()

	var wrote directive.Step
	check := New(&emptySig{empty: true}, func(ctx workflow.Context, stepID string, d directive.Step) error {
		wrote = d
		return nil
	})

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		result, err := check.Run(ctx, &app.WorkflowStep{ID: "step-1"}, &app.Workflow{ID: "wf-1"})
		require.NoError(t, err)
		require.Equal(t, directive.StepSkipGroup, result.Directive)
		require.Equal(t, app.StatusSuccess, result.Status)
		require.Equal(t, "empty_group", result.Reason.Check)
		return nil
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.Equal(t, directive.StepSkipGroup, wrote)
}

func TestEmptyGroupCheckPassesWhenGroupHasInstalls(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()

	check := New(&emptySig{empty: false}, func(ctx workflow.Context, stepID string, d directive.Step) error {
		t.Fatal("should not write a directive for a non-empty group")
		return nil
	})

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		result, err := check.Run(ctx, &app.WorkflowStep{ID: "step-1"}, &app.Workflow{ID: "wf-1"})
		require.NoError(t, err)
		require.Equal(t, directive.StepUnknown, result.Directive)
		return nil
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}
