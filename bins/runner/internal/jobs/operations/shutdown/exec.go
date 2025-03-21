package shutdown

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
)

func (h *handler) finishJob(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	_, err := h.apiClient.UpdateJobExecution(ctx, job.ID, jobExecution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{
		Status: models.AppRunnerJobExecutionStatusFinished,
	})
	if err != nil {
		return err
	}

	shutdownType, ok := job.Metadata["shutdown_type"]
	if ok && shutdownType == "force" {
		if _, err := h.apiClient.UpdateJob(ctx, job.ID, &models.ServiceUpdateRunnerJobRequest{
			Status: models.AppRunnerJobStatusFinished,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("exec", zap.String("job_type", "shutdown"))

	// NOTE(jm): we can not _safely_ stop the fx loop in this step, because the job execution in jobloop.JobLoop
	// will attempt to "cleanup" after this. This means that immediately once this returns, it will set the status
	// of the job execution to cleaning-up, and will expect the cleanup step finishes before updating the job with
	// the real status. This creates a race condition, as we want the shut down to be the very last step of the job,
	// and no updates can happen after it.
	//
	// Thus, we _actually_ execute the shut down in the cleanup.

	return nil
}
