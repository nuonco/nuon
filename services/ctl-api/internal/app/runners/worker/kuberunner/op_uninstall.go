package runner

import (
	"context"
	"fmt"

	"helm.sh/helm/v3/pkg/action"

	"github.com/powertoolsdev/mono/pkg/helm"
)

func (a *Activities) uninstall(ctx context.Context, actionCfg *action.Configuration, runnerID string) error {
	releaseName := fmt.Sprintf("runner-%s", runnerID)
	prevRel, err := helm.GetRelease(actionCfg, releaseName)
	if err != nil {
		return fmt.Errorf("unable to get previous helm release: %w", err)
	}

	if prevRel == nil {
		return nil
	}

	_, err = action.NewUninstall(actionCfg).Run(prevRel.Name)
	if err != nil {
		return fmt.Errorf("unable to uninstall previous release: %w", err)
	}

	return nil
}
