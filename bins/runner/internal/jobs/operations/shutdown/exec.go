package shutdown

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/fx"
)

func (h *handler) finishJob(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	_, err := h.apiClient.UpdateJobExecution(ctx, job.ID, jobExecution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{
		Status: models.AppRunnerJobExecutionStatusFinished,
	})
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if err := h.finishJob(ctx, job, jobExecution); err != nil {
		h.errRecorder.Record("fetch job executions", err)
	}

	if err := h.shutdowner.Shutdown(fx.ExitCode(0)); err != nil {
		h.errRecorder.Record("unable to shut down", err)
	}
	return nil
}
