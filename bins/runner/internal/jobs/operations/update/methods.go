package update

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
)

func (h *handler) Fetch(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	// Ask the API what version this runner should be running
	settings, err := h.apiClient.GetSettings(ctx)

	h.state = &handlerState{}

	if err != nil {
		return errors.Wrap(err, "unable to get settings")
	}
	h.state.expectedVersion = settings.ContainerImageTag
	return nil
}

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("initializing", zap.String("job_type", "update-version"), zap.String("expected_version", h.state.expectedVersion))
	return nil
}

func (h *handler) Validate(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("validating", zap.String("job_type", "update-version"), zap.String("expected_version", h.state.expectedVersion))
	if err := jobs.Matches(job, h); err != nil {
		return err
	}
	return nil
}

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	// As with the shutdown job handler, fx shutdown cannot be safely triggered in this phase.
	// Must be done in cleanup.
	l.Info("exec", zap.String("job_type", "update-version"), zap.String("expected_version", h.state.expectedVersion))
	return nil
}

func (h *handler) Outputs(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (h *handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	// The approach to updating the runner depends on the environment in which
	// it is running.

	// TODO(sdboyer) this should become a big switch that picks the known supervisor version we want.
	// But until we have a strategy other than use-latest, just shut down.

	l.Info("shutting down, supervisor should restart at new version", zap.String("expected_version", h.state.expectedVersion))
	if _, err = h.apiClient.UpdateJobExecution(ctx, job.ID, jobExecution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{
		Status: models.AppRunnerJobExecutionStatusFinished,
	}); err != nil {
		h.errRecorder.Record("update job execution", err)
	}

	if err := h.shutdowner.Shutdown(fx.ExitCode(0)); err != nil {
		h.errRecorder.Record("unable to shut down", err)
	}

	return nil
}
