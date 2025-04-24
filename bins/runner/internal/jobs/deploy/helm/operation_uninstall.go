package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"

	"github.com/powertoolsdev/mono/pkg/helm"
)

func (h *handler) uninstall(ctx context.Context, l *zap.Logger, actionCfg *action.Configuration) error {
	l.Info("fetching previous release")
	prevRel, err := helm.GetRelease(actionCfg, h.state.plan.HelmDeployPlan.Name)
	if err != nil {
		l.Warn("unable to fetch previous release, so assuming it was not installed properly", zap.Error(err))
		return nil
	}

	if prevRel == nil {
		l.Info("no previous release to uninstall")
		return nil
	}

	l.Info("uninstalling release", zap.String("release", prevRel.Name))
	_, err = action.NewUninstall(actionCfg).Run(prevRel.Name)
	if err != nil {
		return fmt.Errorf("unable to uninstall previous release: %w", err)
	}

	return nil
}
