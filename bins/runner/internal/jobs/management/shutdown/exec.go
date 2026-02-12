package shutdown

import (
	"context"
	"fmt"

	"github.com/fidiego/systemctl"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/monitor"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) finishJob(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	_, err := h.apiClient.UpdateJobExecution(ctx, job.ID, jobExecution.ID, &models.ServiceUpdateRunnerJobExecutionRequest{
		Status: models.AppRunnerJobExecutionStatusFinished,
	})
	if err != nil {
		return err
	}

	if _, err := h.apiClient.UpdateJob(ctx, job.ID, &models.ServiceUpdateRunnerJobRequest{
		Status: models.AppRunnerJobStatusFinished,
	}); err != nil {
		return err
	}
	return nil
}

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}
	l.Info("preparing to gracefully shutting down the runner process")

	l.Info("updating image config", zap.String("image_tag", h.settings.ContainerImageTag))
	monitor.EnsureImageConfigFile(ctx, l, h.settings)

	l.Info(fmt.Sprintf("restarting %s", monitor.RunnerServiceName))
	if err := systemctl.Restart(ctx, monitor.RunnerServiceName, systemctl.Options{}); err != nil {
		l.Error("failed to restart runner service", zap.Error(err))
	}

	return nil
}
