package helm

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/pkg/helm"
)

func (h *handler) execUninstall(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if err := h.uninstall(ctx, l, actionCfg); err != nil {
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
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("Initializing Helm...", zapcore.Field{Key: "base_path", Type: zapcore.StringType, String: h.state.arch.BasePath()})
	actionCfg, kubeCfg, err := h.actionInit(ctx, l)
	if err != nil {
		return fmt.Errorf("unable to initialize helm actions: %w", err)
	}
	actionCfg.Log = helm.Logger(l)

	// set the release storage backend dynamically
	releaseStore, err := h.getHelmReleaseStore(ctx, kubeCfg)
        if err != nil {
                return errors.Wrap(err, "unable to get release store")
        }

	actionCfg.Releases = releaseStore

	if job.Operation == models.AppRunnerJobOperationTypeDestroy {
		return h.execUninstall(ctx, l, actionCfg, job, jobExecution)
	}

	l.Info("Checking for previous Helm release...", zapcore.Field{Key: "base_path", Type: zapcore.StringType, String: h.state.arch.BasePath()})
	prevRel, err := helm.GetRelease(actionCfg, h.state.plan.HelmDeployPlan.Name)
	if err != nil {
		return fmt.Errorf("unable to get previous helm release: %w", err)
	}

	var (
		rel *release.Release
		op  string
	)
	if prevRel == nil {
		op = "install"
		rel, err = h.install(ctx, l, actionCfg, kubeCfg)
	} else {
		op = "upgrade"
		rel, err = h.upgrade(ctx, l, actionCfg, kubeCfg)
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
