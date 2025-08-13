package shutdown

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/fx"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
)

func (h *handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("gracefully shutting down the runner process")
	if err := h.finishJob(ctx, job, jobExecution); err != nil {
		h.errRecorder.Record("update job execution", err)
	}

	if err := h.shutdowner.Shutdown(fx.ExitCode(0)); err != nil {
		h.errRecorder.Record("unable to shut down", err)
	}

	return nil
}
