package installstackversionrun

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "install_stack_version_run"
const atomicRunnerStatusVersion = "install-stack-version-run-atomic-runner-status-v1"

type Signal struct {
	RunnerID                 string `json:"runner_id"`
	InstallStackVersionRunID string `json:"install_stack_version_run_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}

	if s.InstallStackVersionRunID == "" {
		return errors.New("install_stack_version_run_id is required")
	}

	// Validate runner exists in database
	_, err := activities.AwaitGetByRunnerID(ctx, s.RunnerID)
	if err != nil {
		return errors.Wrap(err, "runner not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Get runner to check status
	runner, err := activities.AwaitGetByRunnerID(ctx, s.RunnerID)
	if err != nil {
		return err
	}

	// Offline/error are included so a re-applied stack gets a fresh heartbeat window.
	if !generics.SliceContains(runner.Status, []app.RunnerStatus{
		app.RunnerStatusAwaitingInstallStackRun,
		app.RunnerStatusPending,
		app.RunnerStatusOffline,
		app.RunnerStatusError,
	}) {
		return nil
	}

	// A stack applied with runner_enabled = false provisions no runner, so it
	// will never report health. Checking the run's outputs here rather than the
	// runner's status is deliberate: this signal races
	// update-install-stack-outputs off the same stack run, so the disabled
	// status may not be written yet.
	runnerDisabled, err := activities.AwaitGetStackRunRunnerDisabled(ctx, activities.GetStackRunRunnerDisabledRequest{
		InstallStackVersionRunID: s.InstallStackVersionRunID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to determine whether stack run disabled the runner")
	}
	if runnerDisabled {
		return nil
	}

	statusVersion := workflow.GetVersion(ctx, atomicRunnerStatusVersion, workflow.DefaultVersion, 1)
	if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		RunnerID:          s.RunnerID,
		Status:            app.RunnerStatusAwaitingHeartbeat,
		StatusDescription: "runner install stack was run, waiting for the runner to report in",
	}); err != nil {
		return err
	}
	if statusVersion == workflow.DefaultVersion {
		statusactivities.AwaitUpdateRunnerStatusV2(ctx, statusactivities.UpdateRunnerStatusV2Request{
			RunnerID:          s.RunnerID,
			Status:            app.RunnerStatusAwaitingHeartbeat,
			StatusDescription: "runner install stack was run, waiting for the runner to report in",
		})
	}

	return nil
}
