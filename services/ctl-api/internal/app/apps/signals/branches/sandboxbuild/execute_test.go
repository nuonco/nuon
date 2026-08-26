package sandboxbuild

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	branchactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func TestExecuteAddsSandboxBuildIDToStepMetadata(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	// the detector is wall-clock based, so a loaded CI host trips it on a
	// workflow that is not actually blocked
	env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})
	sig := &Signal{
		AppBranchID: "branch-1",
		RunID:       "run-1",
		StepID:      "step-1",
	}

	env.OnActivity((*branchactivities.Activities).GetAppBranchRunByID, mock.Anything, mock.Anything, mock.Anything).
		Return(&app.AppBranchRun{
			AppConfigID: "config-1",
			OrgID:       "org-1",
			CreatedByID: "account-1",
		}, nil)
	env.OnActivity((*branchactivities.Activities).GetAppConfigByID, mock.Anything, mock.Anything, mock.Anything).
		Return(&app.AppConfig{ID: "config-1", AppID: "app-1"}, nil)
	env.OnActivity((*branchactivities.Activities).GetLatestAppSandboxConfig, mock.Anything, mock.Anything, mock.Anything).
		Return(&app.AppSandboxConfig{ID: "sandbox-config-1"}, nil)
	env.OnActivity((*branchactivities.Activities).GetSandboxBuildGitSource, mock.Anything, mock.Anything, mock.Anything).
		Return(&plantypes.GitSource{}, nil)
	env.OnActivity((*branchactivities.Activities).CreateSandboxBuild, mock.Anything, mock.Anything, mock.Anything).
		Return(&app.AppSandboxBuild{ID: "sandbox-build-1"}, nil)
	env.OnActivity((*statusactivities.Activities).PkgStatusUpdateFlowStepStatus, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	var statusUpdate statusactivities.UpdateStatusRequest
	env.SetOnActivityStartedListener(func(info *activity.Info, _ context.Context, args converter.EncodedValues) {
		if info.ActivityType.Name == "PkgStatusUpdateFlowStepStatus" {
			require.NoError(t, args.Get(&statusUpdate))
		}
	})
	env.OnActivity((*branchactivities.Activities).CreateSandboxBuildLogStream, mock.Anything, mock.Anything, mock.Anything).
		Return((*app.LogStream)(nil), temporal.NewNonRetryableApplicationError("log stream unavailable", "TEST", nil))
	env.OnActivity((*branchactivities.Activities).CreateSandboxBuildJob, mock.Anything, mock.Anything, mock.Anything).
		Return((*app.RunnerJob)(nil), temporal.NewNonRetryableApplicationError("stop after metadata update", "TEST", nil))
	env.OnActivity((*branchactivities.Activities).UpdateSandboxBuildStatus, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.Execute(ctx)
	})

	require.ErrorContains(t, env.GetWorkflowError(), "stop after metadata update")
	require.Equal(t, "step-1", statusUpdate.ID)
	require.Equal(t, app.StatusInProgress, statusUpdate.Status.Status)
	require.Equal(t, "sandbox-build-1", statusUpdate.Status.Metadata["sandbox_build_id"])
	env.AssertExpectations(t)
}
