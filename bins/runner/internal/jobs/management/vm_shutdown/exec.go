package shutdown

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	pkgshutdown "github.com/powertoolsdev/mono/bins/runner/internal/pkg/shutdown"
)

func (h *handler) finishJob(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	_, err := h.apiClient.UpdateJobExecution(ctx, job.ID, jobExecution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{
		Status: models.AppRunnerJobExecutionStatusFinished,
	})
	if err != nil {
		return err
	}

	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	shutdownType, ok := job.Metadata["shutdown_type"]
	if ok && shutdownType == "vm" {
		if _, err := h.apiClient.UpdateJob(ctx, job.ID, &models.ServiceUpdateRunnerJobRequest{
			Status: models.AppRunnerJobStatusFinished,
		}); err != nil {
			return err
		}

		// NOTE(fd): this shuts down so quickly we do lose the tail end of the logs.
		// executes an os shutdown ↴ via dbus w/ a shell fallback w/ a sudo shell fallback
		err = pkgshutdown.Shutdown(ctx, l, h.v)
		if err != nil {
		}
	}

	return nil
}

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	// 1. shutdown the runner systemd service
	// 2. send shutdown signal to the VM w/ enough of a delay for this process to finish (cleanup)
	// 3. TODO: in cleanup, consider stopping the `runner mng` systemd (although we do lose recoverability in case of shutdown failure)

	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("exec", zap.String("job_type", "shutdown"))

	// NOTE: this job shuts down the whole VM so we execute the work from within the cleanup. see `finishJob`.

	return nil
}
