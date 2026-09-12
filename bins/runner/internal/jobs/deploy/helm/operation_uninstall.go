package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/action"
	kube "helm.sh/helm/v4/pkg/kube"

	"github.com/nuonco/nuon/pkg/helm"
)

func (h *handler) uninstall(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration) error {
	releaseName := h.state.plan.HelmDeployPlan.Name

	l.Info("fetching previous release")
	// Error before nil: a store failure read as "not installed" would report a
	// successful teardown while the release and its resources stayed in place.
	prevRel, err := helm.GetRelease(actionCfg, releaseName)
	if err != nil {
		return fmt.Errorf("unable to read release %s before uninstalling it: %w", releaseName, err)
	}

	if prevRel == nil {
		l.Info("no previous release to uninstall", zap.String("release", releaseName))
		return nil
	}

	l.Info("uninstalling release", zap.String("release", prevRel.Name))
	client := action.NewUninstall(actionCfg)
	// NOTE(fd): determine what the right wait strategy should be here
	client.WaitStrategy = kube.StatusWatcherStrategy
	client.Timeout = h.state.timeout
	// The release can go away between the read above and the purge below, either
	// from a concurrent teardown or from a store that 404s the record it served.
	client.IgnoreNotFound = true

	if _, err := client.Run(prevRel.Name); err != nil {
		if confirmErr := helm.ConfirmUninstalled(actionCfg, prevRel.Name, err); confirmErr != nil {
			return fmt.Errorf("unable to uninstall previous release: %w", confirmErr)
		}

		l.Warn("helm reported the release was not found while uninstalling it, and it is no longer stored, so the uninstall is complete",
			zap.String("release", prevRel.Name),
			zap.Error(err),
		)
	}

	return nil
}
