package runner

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/action"
	"helm.sh/helm/v4/pkg/kube"

	"github.com/nuonco/nuon/pkg/helm"
)

func (a *Activities) uninstall(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration, runnerID string) error {
	releaseName := fmt.Sprintf("runner-%s", runnerID)
	prevRel, err := helm.GetRelease(actionCfg, releaseName)
	if err != nil {
		return fmt.Errorf("unable to get previous helm release: %w", err)
	}

	if prevRel == nil {
		return nil
	}

	client := action.NewUninstall(actionCfg)
	client.WaitStrategy = kube.StatusWatcherStrategy
	client.Timeout = defaultHelmOperationTimeout
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
