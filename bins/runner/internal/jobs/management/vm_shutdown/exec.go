package shutdown

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	pkgshutdown "github.com/nuonco/nuon/bins/runner/internal/pkg/shutdown"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
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

		// why: On Azure, don't power the VM off. An instance refresh in Azure takes 10m+ to complete.
		// Keep the VM on and let the Azure control plane replace it.
		if h.settings.Platform == "azure" {
			l.Info("vm shutdown - marking vm as unhealthy; letting azure vmss replace the instance")
			h.health.SetUnhealthy()
			return nil
		}

		if err := pkgshutdown.Shutdown(ctx, l, h.v); err != nil {
			l.Error("failed to shut down vm", zap.Error(err))
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

	return nil
}
