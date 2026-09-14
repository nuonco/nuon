package generateinstallstackversion

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	pkgstate "github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

// TestPrepareRegeneratesRunnerStateAfterInstallCreatedState pins the ordering of
// the two state regenerations prepare runs.
//
// A regeneration reads the latest state, fetches only its own partials and saves
// a new row. Running the install-created and runner regenerations concurrently
// therefore makes the last writer drop the other's partials from the current
// state — observed as an install whose newest state had only the runner partial
// populated, losing org, app, cloud_account and inputs.
func (s *SignalTestSuite) TestPrepareRegeneratesRunnerStateAfterInstallCreatedState() {
	env := s.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})

	var order []string
	var saveTriggers []string
	env.SetOnActivityStartedListener(func(info *activity.Info, _ context.Context, args converter.EncodedValues) {
		name := info.ActivityType.Name
		order = append(order, name)
		if name == "SaveState" {
			var req activities.SaveStateRequest
			if err := args.Get(&req); err == nil {
				saveTriggers = append(saveTriggers, req.TriggeredByType)
			}
		}
	})

	install := &app.Install{
		ID:          "inst-1",
		Name:        "test-install",
		AppID:       "app-1",
		AppConfigID: "cfg-1",
		RunnerID:    "run-1",
	}

	env.OnActivity((*sharedactivities.Activities).EnqueueSignalToOwner, mock.Anything, mock.Anything, mock.Anything).
		Return(&sharedactivities.EnqueueSignalToOwnerResponse{}, nil).Once()
	env.OnActivity((*activities.Activities).GetLatestInstallState, mock.Anything, mock.Anything, mock.Anything).
		Return(pkgstate.New(), nil).Maybe()
	env.OnActivity((*activities.Activities).Get, mock.Anything, mock.Anything, mock.Anything).
		Return(install, nil).Maybe()
	env.OnActivity((*activities.Activities).GetOrg, mock.Anything, mock.Anything, mock.Anything).
		Return(&app.Org{ID: "org-1", Name: "test-org"}, nil).Maybe()
	env.OnActivity((*activities.Activities).GetRunner, mock.Anything, mock.Anything, mock.Anything).
		Return(&app.Runner{ID: "run-1"}, nil).Maybe()
	env.OnActivity((*activities.Activities).GetInstallInputsState, mock.Anything, mock.Anything, mock.Anything).
		Return(&app.InstallInputs{ID: "inp-1"}, nil).Maybe()
	env.OnActivity((*activities.Activities).GetAppConfigInputSection, mock.Anything, mock.Anything, mock.Anything).
		Return(&app.AppConfig{ID: "cfg-1"}, nil).Maybe()
	env.OnActivity((*activities.Activities).SaveState, mock.Anything, mock.Anything, mock.Anything).
		Return(&app.InstallState{ID: "st-1"}, nil).Twice()
	env.OnActivity((*activities.Activities).ArchiveState, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Maybe()
	env.OnActivity((*activities.Activities).RenderInstallLabels, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Maybe()

	sig := &Signal{}
	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.prepare(ctx, install)
	})

	require.NoError(s.T(), env.GetWorkflowError())
	require.Equal(s.T(), []string{"installs", "runners"}, saveTriggers,
		"install-created state must be saved before the runner state")

	firstSave := indexAfter(order, "SaveState", -1)
	require.GreaterOrEqual(s.T(), firstSave, 0, "expected a SaveState call")
	require.Greater(s.T(), indexAfter(order, "GetLatestInstallState", firstSave), firstSave,
		"runner regeneration must read its base state after the install-created state was saved, "+
			"otherwise it overwrites the current state with only the runner partial")
}

func indexAfter(names []string, want string, after int) int {
	for i := after + 1; i < len(names); i++ {
		if names[i] == want {
			return i
		}
	}
	return -1
}

func TestIndexAfter(t *testing.T) {
	names := []string{"a", "b", "a", "c"}
	require.Equal(t, 0, indexAfter(names, "a", -1))
	require.Equal(t, 2, indexAfter(names, "a", 0))
	require.Equal(t, -1, indexAfter(names, "z", -1))
	require.Equal(t, -1, indexAfter(names, "a", 2))
}
