package installstackversionrun

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "install_stack_version_run"

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

	// Only update status if runner is in specific states
	if !generics.SliceContains(runner.Status, []app.RunnerStatus{
		app.RunnerStatusAwaitingInstallStackRun,
		app.RunnerStatusPending,
	}) {
		return nil
	}

	// Update runner status to Error state
	// This indicates the install stack was run and is waiting for health check
	if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		RunnerID:          s.RunnerID,
		Status:            app.RunnerStatusError,
		StatusDescription: "runner install stack was run, waiting for health check to mark healthy",
	}); err != nil {
		return err
	}

	return nil
}
