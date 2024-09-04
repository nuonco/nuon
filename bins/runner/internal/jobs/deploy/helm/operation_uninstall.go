package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"

	"github.com/powertoolsdev/mono/pkg/helm"
)

func (h *handler) uninstall(ctx context.Context, actionCfg *action.Configuration) error {
	h.log.Info("fetching previous release")
	prevRel, err := helm.GetRelease(actionCfg, h.state.cfg.Name)
	if err != nil {
		return fmt.Errorf("unable to get previous helm release: %w", err)
	}

	if prevRel == nil {
		h.log.Info("no previous release to uninstall")
		return nil
	}

	h.log.Info("uninstalling release", zap.String("release", prevRel.Name))
	_, err = action.NewUninstall(actionCfg).Run(prevRel.Name)
	if err != nil {
		return fmt.Errorf("unable to uninstall previous release: %w", err)
	}

	return nil
}
