package shutdown

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
)

func (h *handler) Cleanup(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("shutting down the vm")
	if err := h.finishJob(ctx, job, jobExecution); err != nil {
		h.errRecorder.Record("update job execution", err)
	}

	return nil
}
