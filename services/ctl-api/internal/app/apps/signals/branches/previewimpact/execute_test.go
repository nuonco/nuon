package previewimpact

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	branchactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
)

func TestPreviewImpactNoConfigChanges(t *testing.T) {
	for _, tc := range []struct {
		name            string
		noConfigChanges bool
		force           bool
		appConfigID     string
		wantErr         string
	}{
		{name: "nothing to preview is skipped", noConfigChanges: true, appConfigID: ""},
		{name: "forced run still requires a config", noConfigChanges: true, force: true, appConfigID: "", wantErr: "has no app config ID"},
		{name: "missing config without the flag is still an error", noConfigChanges: false, appConfigID: "", wantErr: "has no app config ID"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment()
			env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})

			env.OnActivity((*branchactivities.Activities).GetAppBranchRunByID, mock.Anything, mock.Anything, mock.Anything).
				Return(&app.AppBranchRun{
					ID:              "run-1",
					NoConfigChanges: tc.noConfigChanges,
					Force:           tc.force,
					AppConfigID:     tc.appConfigID,
				}, nil)

			sig := &Signal{RunID: "run-1", AppBranchID: "branch-1", AppBranchConfigID: "branch-config-1"}
			env.ExecuteWorkflow(sig.Execute)

			require.True(t, env.IsWorkflowCompleted())
			if tc.wantErr == "" {
				require.NoError(t, env.GetWorkflowError())
			} else {
				require.ErrorContains(t, env.GetWorkflowError(), tc.wantErr)
			}
		})
	}
}
