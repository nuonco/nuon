package helm

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"

	"github.com/powertoolsdev/mono/pkg/helm"
)

func (h *handler) execUninstall(ctx context.Context, actionCfg *action.Configuration, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if err := h.uninstall(ctx, actionCfg); err != nil {
		h.writeErrorResult(ctx, "uninstall", err)
		return fmt.Errorf("unable to uninstall helm chart: %w", err)
	}

	res := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success: true,
	}
	if _, err := h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, res); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}

	return nil
}

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	h.log.Info("Initializing Helm...")
	helmClient, err := h.actionInit(ctx, h.log)
	if err != nil {
		return fmt.Errorf("unable to initialize helm actions: %w", err)
	}

	if job.Operation == models.AppRunnerJobOperationTypeDestroy {
		return h.execUninstall(ctx, helmClient, job, jobExecution)
	}

	h.log.Info("Checking for previous Helm release...")
	prevRel, err := helm.GetRelease(helmClient, h.state.cfg.Name)
	if err != nil {
		return fmt.Errorf("unable to get previous helm release: %w", err)
	}

	var (
		rel *release.Release
		op  string
	)
	if prevRel == nil {
		op = "install"
		rel, err = h.install(ctx, helmClient)
	} else {
		op = "upgrade"
		rel, err = h.upgrade(ctx, helmClient)
	}
	if err != nil {
		h.writeErrorResult(ctx, op, err)
		return fmt.Errorf("unable to %s helm chart: %w", op, err)
	}

	apiRes, err := h.createAPIResult(rel)
	if err != nil {
		h.writeErrorResult(ctx, op, err)
		return fmt.Errorf("unable to create api result from release: %w", err)
	}
	if _, err := h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, apiRes); err != nil {
		h.errRecorder.Record("write job execution result", err)
	}

	return nil
}
