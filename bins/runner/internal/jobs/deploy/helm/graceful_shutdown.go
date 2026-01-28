package helm

import (
	"context"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/action"
	release "helm.sh/helm/v4/pkg/release/v1"

	"github.com/nuonco/nuon/pkg/helm"
)

const gracefulShutdownTimeout = 30 * time.Second

func (h *handler) GracefulShutdown(ctx context.Context, job *models.AppRunnerJob, l *zap.Logger) error {
	if h.state == nil {
		return nil
	}

	// Only cleanup for failed/timed-out/cancelled jobs
	switch job.Status {
	case models.AppRunnerJobStatusFailed,
		models.AppRunnerJobStatusTimedDashOut,
		models.AppRunnerJobStatusCancelled:
		// Proceed with cleanup
	default:
		// Job finished successfully or never started - no cleanup needed
		l.Debug("skipping helm graceful shutdown - job status does not require cleanup",
			zap.String("status", string(job.Status)))
		return nil
	}

	// Check if we have the necessary plan information
	if h.state.plan == nil || h.state.plan.HelmDeployPlan == nil {
		l.Debug("skipping helm graceful shutdown - no helm deploy plan available")
		return nil
	}

	l.Info("attempting helm graceful shutdown",
		zap.String("release_name", h.state.plan.HelmDeployPlan.Name),
		zap.String("namespace", h.state.plan.HelmDeployPlan.Namespace))

	// Initialize helm action configuration
	actionCfg, _, err := h.actionInit(ctx, l)
	if err != nil {
		h.writeErrorResult(ctx, "helm graceful shutdown: init action config", err)
		l.Warn("failed to initialize helm action config during graceful shutdown", zap.Error(err))
		return nil // Don't fail graceful shutdown
	}

	// Get current release to check its status
	currentRelease, err := helm.GetRelease(actionCfg, h.state.plan.HelmDeployPlan.Name)
	if err != nil {
		h.writeErrorResult(ctx, "helm graceful shutdown: get release", err)
		l.Warn("failed to get helm release during graceful shutdown", zap.Error(err))
		return nil // Don't fail graceful shutdown
	}

	if currentRelease == nil {
		l.Debug("no helm release found during graceful shutdown - nothing to cleanup")
		return nil
	}

	l.Info("found helm release during graceful shutdown",
		zap.String("release_name", currentRelease.Name),
		zap.String("status", currentRelease.Info.Status.String()),
		zap.Int("version", currentRelease.Version))

	// Handle based on release status
	switch currentRelease.Info.Status {
	case release.StatusPendingInstall:
		// No previous version to rollback to - must uninstall
		l.Info("release is in pending-install state, attempting uninstall")
		if err := h.gracefulUninstall(ctx, l, actionCfg, currentRelease.Name); err != nil {
			h.writeErrorResult(ctx, "helm graceful shutdown: uninstall pending-install", err)
			l.Warn("failed to uninstall pending-install release during graceful shutdown", zap.Error(err))
		} else {
			l.Info("successfully uninstalled pending-install release during graceful shutdown")
		}

	case release.StatusPendingUpgrade, release.StatusPendingRollback, release.StatusUninstalling:
		// Rollback to last known good state
		l.Info("release is in pending state, attempting rollback to last deployed version",
			zap.String("status", currentRelease.Info.Status.String()))
		if err := h.gracefulRollback(ctx, l, actionCfg, currentRelease.Name); err != nil {
			h.writeErrorResult(ctx, "helm graceful shutdown: rollback", err)
			l.Warn("failed to rollback release during graceful shutdown", zap.Error(err))
		} else {
			l.Info("successfully rolled back release during graceful shutdown")
		}

	default:
		// Release is in a stable state (deployed, failed, superseded, etc.)
		l.Debug("release is in stable state, no graceful shutdown action needed",
			zap.String("status", currentRelease.Info.Status.String()))
	}

	return nil
}

func (h *handler) gracefulUninstall(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration, releaseName string) error {
	uninstall := action.NewUninstall(actionCfg)
	uninstall.Timeout = gracefulShutdownTimeout
	uninstall.DisableHooks = true // Avoid hanging on hooks during shutdown
	uninstall.KeepHistory = false // Clean up completely

	l.Debug("executing helm uninstall during graceful shutdown",
		zap.String("release_name", releaseName),
		zap.Duration("timeout", gracefulShutdownTimeout))

	_, err := uninstall.Run(releaseName)
	return err
}

func (h *handler) gracefulRollback(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration, releaseName string) error {
	// First, find the last deployed version
	lastDeployedVersion, err := helm.FindLastDeployedVersion(actionCfg, releaseName)
	if err != nil {
		l.Warn("failed to find last deployed version, falling back to uninstall", zap.Error(err))
		return h.gracefulUninstall(ctx, l, actionCfg, releaseName)
	}

	if lastDeployedVersion == 0 {
		l.Info("no previously deployed version found, falling back to uninstall")
		return h.gracefulUninstall(ctx, l, actionCfg, releaseName)
	}

	l.Debug("found last deployed version", zap.Int("version", lastDeployedVersion))

	rollback := action.NewRollback(actionCfg)
	rollback.Version = lastDeployedVersion
	rollback.Timeout = gracefulShutdownTimeout
	rollback.DisableHooks = true // Avoid hanging on hooks during shutdown
	rollback.CleanupOnFail = true

	l.Debug("executing helm rollback during graceful shutdown",
		zap.String("release_name", releaseName),
		zap.Int("target_version", lastDeployedVersion),
		zap.Duration("timeout", gracefulShutdownTimeout))

	return rollback.Run(releaseName)
}
