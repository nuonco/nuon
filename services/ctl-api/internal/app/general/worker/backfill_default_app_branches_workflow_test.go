package worker

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/types/workflows/defaultappbranches"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
)

// Activities are registered by name before any mock is set up: the test
// environment only deserializes an activity's input for mocks when the activity
// itself is registered, and every registration has to precede the first
// OnActivity call.
func newBackfillDefaultAppBranchesEnv(t *testing.T, appIDs []string) *testsuite.TestWorkflowEnvironment {
	t.Helper()

	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})

	acts := &activities.Activities{}
	env.RegisterActivityWithOptions(acts.ListAppsNeedingDefaultBranch,
		activity.RegisterOptions{Name: "ListAppsNeedingDefaultBranch"})
	env.RegisterActivityWithOptions(acts.EnsureDefaultAppBranch,
		activity.RegisterOptions{Name: "EnsureDefaultAppBranch"})

	env.OnActivity("ListAppsNeedingDefaultBranch", mock.Anything, mock.Anything).
		Return(&activities.ListAppsNeedingDefaultBranchResponse{AppIDs: appIDs}, nil).Once()

	return env
}

func backfillProgress(t *testing.T, env *testsuite.TestWorkflowEnvironment) defaultappbranches.Progress {
	t.Helper()

	encoded, err := env.QueryWorkflow(defaultappbranches.ProgressQueryType)
	require.NoError(t, err)

	var progress defaultappbranches.Progress
	require.NoError(t, encoded.Get(&progress))
	return progress
}

func TestBackfillDefaultAppBranchesDryRunCreatesNothing(t *testing.T) {
	env := newBackfillDefaultAppBranchesEnv(t, []string{"app-1", "app-2", "app-3"})

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return (&Workflows{}).BackfillDefaultAppBranches(ctx, defaultappbranches.Request{DryRun: true})
	})
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	progress := backfillProgress(t, env)
	require.True(t, progress.DryRun)
	require.True(t, progress.Done)
	require.Equal(t, 3, progress.AppsTotal)
	require.Equal(t, 0, progress.AppsDone, "a dry run must not ensure any branch")
	env.AssertExpectations(t)
}

// A fleet-wide backfill that stops at the first bad app leaves the rest
// unmigrated, so a failure is counted and the run carries on.
func TestBackfillDefaultAppBranchesTalliesOutcomes(t *testing.T) {
	env := newBackfillDefaultAppBranchesEnv(t, []string{"app-new", "app-existing", "app-claimed", "app-broken"})

	outcomes := map[string]string{
		"app-new":      activities.DefaultAppBranchOutcomeCreated,
		"app-existing": activities.DefaultAppBranchOutcomeExists,
		"app-claimed":  activities.DefaultAppBranchOutcomeClaimed,
	}
	for appID, outcome := range outcomes {
		env.OnActivity("EnsureDefaultAppBranch", mock.Anything,
			activities.EnsureDefaultAppBranchRequest{AppID: appID}).
			Return(&activities.EnsureDefaultAppBranchResponse{AppID: appID, Outcome: outcome, InstallsConnected: 2}, nil).Once()
	}
	env.OnActivity("EnsureDefaultAppBranch", mock.Anything,
		activities.EnsureDefaultAppBranchRequest{AppID: "app-broken"}).
		Return(nil, errors.New("boom")).Times(defaultAppBranchAttempts)

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return (&Workflows{}).BackfillDefaultAppBranches(ctx, defaultappbranches.Request{})
	})
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	progress := backfillProgress(t, env)
	require.True(t, progress.Done)
	require.Equal(t, 4, progress.AppsTotal)
	require.Equal(t, 4, progress.AppsDone)
	require.Equal(t, 1, progress.Created)
	require.Equal(t, 1, progress.Existing)
	require.Equal(t, 1, progress.Claimed)
	require.Equal(t, 1, progress.Failed)
	require.Equal(t, []string{"app-broken"}, progress.FailedAppIDs)
	require.Equal(t, 6, progress.InstallsConnected, "installs connected are summed across outcomes, not just new branches")
	env.AssertExpectations(t)
}

func TestBackfillDefaultAppBranchesContinuesAsNewPastRunCap(t *testing.T) {
	appIDs := make([]string, defaultAppBranchesPerRun+1)
	for i := range appIDs {
		appIDs[i] = fmt.Sprintf("app-%d", i)
	}

	env := newBackfillDefaultAppBranchesEnv(t, appIDs)
	env.OnActivity("EnsureDefaultAppBranch", mock.Anything, mock.Anything).
		Return(&activities.EnsureDefaultAppBranchResponse{Outcome: activities.DefaultAppBranchOutcomeCreated}, nil).
		Times(defaultAppBranchesPerRun)

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return (&Workflows{}).BackfillDefaultAppBranches(ctx, defaultappbranches.Request{})
	})
	require.True(t, env.IsWorkflowCompleted())

	var continueErr *workflow.ContinueAsNewError
	require.ErrorAs(t, env.GetWorkflowError(), &continueErr,
		"the run cap must hand the remainder to a new run, not drop it")
	env.AssertExpectations(t)
}
