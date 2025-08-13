package update

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/fx"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
)

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
